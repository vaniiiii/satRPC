package task

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/satlayer/satlayer-api/chainio/types"
	"github.com/satlayer/satlayer-api/logger"
	transactionprocess "github.com/satlayer/satlayer-api/metrics/indicators/transaction_process"

	"github.com/satlayer/satlayer-api/chainio/io"

	BvsSquaringApi "github.com/satlayer/hello-world-bvs/bvs_squaring_api"
	"github.com/satlayer/hello-world-bvs/task/core"
	"github.com/satlayer/satlayer-api/chainio/api"
)

type Caller struct {
	bvsContract string
	chainIO     io.ChainIO
}

// RunCaller runs the caller by creating a new caller and executing its Run method.
//
// No parameters.
// No return.
func RunCaller() {
	c := NewCaller()
	c.Run()
}

// NewCaller creates a new Caller instance.
//
// Returns a pointer to Caller.
func NewCaller() *Caller {
	elkLogger := logger.NewELKLogger("bvs_demo")
	elkLogger.SetLogLevel("info")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, core.C.Owner.KeyDir, core.C.Owner.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             5,
		RetryInterval:          3 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	client, err := chainIO.SetupKeyring(core.C.Owner.KeyName, core.C.Owner.KeyringBackend)
	if err != nil {
		panic(err)
	}
	txResp, err := api.NewBVSDirectoryImpl(client, core.C.Chain.BvsDirectory).GetBVSInfo(core.C.Chain.BvsHash)
	fmt.Printf("txResp: %+v\n", txResp)
	if err != nil {
		panic(err)
	}
	return &Caller{
		bvsContract: txResp.BVSContract,
		chainIO:     client,
	}
}

// Run runs the caller in an infinite loop, creating a new task for a specified operator address every 30 seconds.
//
// No parameters.
// No return.
func (c *Caller) Run() {
	bvsSquaring := BvsSquaringApi.NewBVSSquaring(c.chainIO)
	operator := "bbn1d9878dze7npzf7t3vxh8f5y2munj7a8xuy50m8"

	for {
		bvsSquaring.BindClient(c.bvsContract)
		resp, err := bvsSquaring.CreateNewTask(context.Background(), operator)
		if err != nil {
			fmt.Printf("Error creating task for operator %s: %v\n", operator, err)
			continue
		}
		fmt.Printf("Created task for operator %s with tx hash: %s\n", operator, resp.Hash.String())

		time.Sleep(30 * time.Second)
	}
}
