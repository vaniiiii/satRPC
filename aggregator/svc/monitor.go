package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/satlayer/satlayer-api/chainio/types"
	"github.com/satlayer/satlayer-api/logger"
	transactionprocess "github.com/satlayer/satlayer-api/metrics/indicators/transaction_process"

	"github.com/satlayer/satlayer-api/chainio/io"

	"github.com/satlayer/hello-world-bvs/aggregator/core"
	BvsSquaringApi "github.com/satlayer/hello-world-bvs/bvs_squaring_api"
	"github.com/satlayer/satlayer-api/chainio/api"
)

var MONITOR Monitor

func init() {
	MONITOR = *NewMonitor()
}

type Monitor struct {
	bvsContract     string
	bvsDirectoryApi api.BVSDirectory
	chainIO         io.ChainIO
}

// NewMonitor creates a new Monitor instance with a Cosmos client and BVS contract.
//
// It takes no parameters.
// Returns a pointer to a Monitor struct.
func NewMonitor() *Monitor {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot get current file")
	}

	configDir := filepath.Dir(currentFile)
	homeDir := configDir + "/../../" + core.C.Owner.KeyDir
	fmt.Printf("homeDir: %s\n", homeDir)
	elkLogger := logger.NewELKLogger("bvs_demo")
	elkLogger.SetLogLevel("info")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, homeDir, core.C.Owner.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             5,
		RetryInterval:          3 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	chainIO, err = chainIO.SetupKeyring(core.C.Owner.KeyName, core.C.Owner.KeyringBackend)
	if err != nil {
		panic(err)
	}
	account, err := chainIO.GetCurrentAccount()
	if err != nil {
		panic(err)
	}
	address := account.GetAddress().String()
	fmt.Printf("address: %s\n", address)
	txResp, err := api.NewBVSDirectoryImpl(chainIO, core.C.Chain.BvsDirectory).GetBVSInfo(core.C.Chain.BvsHash)
	if err != nil {
		panic(err)
	}
	bvsDirectoryApi := api.NewBVSDirectoryImpl(chainIO, core.C.Chain.BvsDirectory)

	return &Monitor{
		bvsContract:     txResp.BVSContract,
		bvsDirectoryApi: bvsDirectoryApi,
		chainIO:         chainIO,
	}
}

// Run starts the task queue monitoring process.
//
// It takes a context.Context object as a parameter.
// No return values.
func (m *Monitor) Run(ctx context.Context) {
	core.L.Info("Start to monitor task queue")
	for {
		results, err := core.S.RedisConn.BLPop(context.Background(), 0, core.PkTaskQueue).Result()
		fmt.Printf("results: %+v\n", results)
		if err != nil {
			core.L.Error(fmt.Sprintf("Failed to read task queue, due to {%s}", err))
			continue
		}
		fmt.Printf("result--->: %s\n", results[1])
		task := core.Task{}
		if err := json.Unmarshal([]byte(results[1]), &task); err != nil {
			core.L.Error(fmt.Sprintf("Failed to parse task queue, due to {%s}", err))
			continue
		}
		fmt.Printf("task: %+v\n", task)
		pkTaskResult := fmt.Sprintf("%s%d", core.PkTaskResult, task.TaskId)
		taskResultStr, err := json.Marshal(task.TaskResult)
		if err != nil {
			core.L.Error(fmt.Sprintf("Failed to marshal task result, due to {%s}", err))
			return
		}
		if err := core.S.RedisConn.LPush(ctx, pkTaskResult, taskResultStr).Err(); err != nil {
			core.L.Error(fmt.Sprintf("Failed to save task result, due to {%s}", err))
			return
		}
		m.verifyTask(ctx, task.TaskId)
	}
}

// verifyTask is a method of the Monitor struct. It is responsible for verifying a task
// by reading the task result from Redis and checking if the result has reached a
// certain threshold. If the threshold is met, it sets the task as finished in Redis,
// logs the task result and operators, and sends the task result to a specified
// destination.
//
// The function takes a context.Context object and a uint64 taskId as parameters.
// It does not return anything.
func (m *Monitor) verifyTask(ctx context.Context, taskId uint64) {
	pkTaskResult := fmt.Sprintf("%s%d", core.PkTaskResult, taskId)
	// timer to read redis queue and verify the task result
	results, err := core.S.RedisConn.LRange(ctx, pkTaskResult, 0, -1).Result()
	fmt.Printf("verify results: %s\n", results)
	if err != nil {
		core.L.Error(fmt.Sprintf("Failed to read task result, due to {%s}", err))
		return
	}

	resultCntMap := make(map[int64]uint)
	resultOperatorMap := make(map[int64][]string)
	var taskResult core.TaskResult
	for _, result := range results {
		fmt.Printf("verify result: %s\n", result)
		if err := json.Unmarshal([]byte(result), &taskResult); err != nil {
			core.L.Error(fmt.Sprintf("Failed to parse task result, due to {%s}", err))
			return
		}
		resultCntMap[taskResult.Result]++
		resultOperatorMap[taskResult.Result] = append(resultOperatorMap[taskResult.Result], taskResult.Operator)
		if resultCntMap[taskResult.Result] >= core.C.App.Threshold {
			pkTaskFinished := fmt.Sprintf("%s%d", core.PkTaskFinished, taskId)
			if err := core.S.RedisConn.Set(ctx, pkTaskFinished, taskResult.Result, 0).Err(); err != nil {
				core.L.Error(fmt.Sprintf("Failed to set task finished, due to {%s}", err))
				return
			}
			operators := strings.Join(resultOperatorMap[taskResult.Result], "&")
			core.L.Info(fmt.Sprintf("Task {%d} is finished. The result is {%d}. The operators are {%s}", taskId, taskResult.Result, operators))
			if err := m.sendTaskResult(taskId, taskResult.Result, operators); err != nil {
				core.L.Error(fmt.Sprintf("Failed to send task result, due to {%s}", err))
			}
			pkTaskOperator := fmt.Sprintf("%s%d", core.PkTaskOperator, taskId)
			core.S.RedisConn.Del(ctx, pkTaskResult)
			core.S.RedisConn.Del(ctx, pkTaskOperator)
			break
		}
	}
}

// sendTaskResult sends the task result to BVS Squaring API.
//
// taskId: the unique identifier of the task
// result: the result of the task
// operators: the operators involved in the task
// error: an error if the task result sending fails
func (m *Monitor) sendTaskResult(taskId uint64, result int64, operators string) error {
	fmt.Println("sendTaskResult", taskId, result, operators)

	bvsSquaring := BvsSquaringApi.NewBVSSquaring(m.chainIO)
	bvsSquaring.BindClient(m.bvsContract)
	_, err := bvsSquaring.RespondToTask(context.Background(), int64(taskId), result, operators)
	if err != nil {
		return err
	}

	return nil
}

func (m *Monitor) VerifyOperator(operator string) (bool, error) {
	rsp, err := m.bvsDirectoryApi.QueryOperator(operator, operator)
	if err != nil {
		core.L.Error(fmt.Sprintf("Failed to query operator, due to {%s}", err))
		return false, err
	}
	fmt.Printf("txnRsp: %+v\n", rsp)
	if rsp.Status == "registered" {
		return true, nil
	}

	return false, nil
}
