package test

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/satlayer/satlayer-api/chainio/types"
	"github.com/satlayer/satlayer-api/logger"
	transactionprocess "github.com/satlayer/satlayer-api/metrics/indicators/transaction_process"

	"github.com/satlayer/satlayer-api/chainio/io"

	BvsSquaringApi "github.com/satlayer/hello-world-bvs/bvs_squaring_api"
	"github.com/stretchr/testify/assert"
)

func testExecuteSquaring(t *testing.T) {
	contrAddr := "bbn1mzq6xzynh002x090rzt6us37scfexpu8ca4sllc3p3scn5mvsp0q5cs73s"
	chainID := "sat-bbn-testnet1"
	rpcURI := "https://rpc.sat-bbn-testnet1.satlayer.net"
	homeDir := "../../.babylond"
	keyName := "wallet1"

	t.Logf("TestExecuteSquaring")
	elkLogger := logger.NewMockELKLogger()
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(chainID, rpcURI, homeDir, "bbn", elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             3,
		RetryInterval:          1 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	assert.NoError(t, err, "failed to create chain IO")
	chainIO, err = chainIO.SetupKeyring(keyName, "test")
	assert.NoError(t, err, "failed to setup keyring")

	bvsSquaring := BvsSquaringApi.NewBVSSquaring(chainIO)
	bvsSquaring.BindClient(contrAddr)

	resp, err := bvsSquaring.CreateNewTask(context.Background(), 10)
	assert.NoError(t, err, "execute contract")
	assert.NotNil(t, resp, "response nil")
	t.Logf("resp:%+v", resp)

	resp, err = bvsSquaring.RespondToTask(context.Background(), 10, 100, "bbn1rt6v30zxvhtwet040xpdnhz4pqt8p2za7y430x")
	assert.NoError(t, err, "execute contract")
	assert.NotNil(t, resp, "response nil")
	t.Logf("resp:%+v", resp)
}

func testQuerySquaring(t *testing.T) {
	contrAddr := "bbn1mzq6xzynh002x090rzt6us37scfexpu8ca4sllc3p3scn5mvsp0q5cs73s"
	chainID := "sat-bbn-testnet1"
	rpcURI := "https://rpc.sat-bbn-testnet1.satlayer.net"
	homeDir := "../../.babylond"
	keyName := "wallet1"

	elkLogger := logger.NewMockELKLogger()
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(chainID, rpcURI, homeDir, "bbn", elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             3,
		RetryInterval:          1 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	assert.NoError(t, err, "failed to create chain IO")
	chainIO, err = chainIO.SetupKeyring(keyName, "test")
	assert.NoError(t, err, "failed to setup keyring")

	bvsSquaring := BvsSquaringApi.NewBVSSquaring(chainIO)
	bvsSquaring.BindClient(contrAddr)

	resp, err := bvsSquaring.GetTaskInput(10)
	assert.NoError(t, err, "execute contract")
	assert.NotNil(t, resp, "response nil")
	t.Logf("resp:%+v", resp)

	resp, err = bvsSquaring.GetTaskResult(10)
	assert.NoError(t, err, "execute contract")
	assert.NotNil(t, resp, "response nil")
	t.Logf("resp:%+v", resp)
}
