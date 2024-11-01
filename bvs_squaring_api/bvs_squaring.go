package BvsSquaringApi

import (
	"context"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/satlayer/satlayer-api/chainio/io"
	"github.com/satlayer/satlayer-api/chainio/types"
)

type BVSSquaring interface {
	BindClient(string)
	CreateNewTask(context.Context, int64) (*coretypes.ResultTx, error)
	RespondToTask(ctx context.Context, taskId int64, result int64, operators string) (*coretypes.ResultTx, error)
	GetTaskInput(int64) (*wasmtypes.QuerySmartContractStateResponse, error)
	GetTaskResult(int64) (*wasmtypes.QuerySmartContractStateResponse, error)
}

type bvsSquaringImpl struct {
	io             io.ChainIO
	executeOptions *types.ExecuteOptions
	queryOptions   *types.QueryOptions
}

func (a *bvsSquaringImpl) BindClient(contractAddress string) {
	a.executeOptions = &types.ExecuteOptions{
		ContractAddr:  contractAddress,
		ExecuteMsg:    []byte{},
		Funds:         "",
		GasAdjustment: 1.2,
		GasPrice:      sdktypes.NewInt64DecCoin("ubbn", 1),
		Gas:           300000,
		Memo:          "test tx",
		Simulate:      true,
	}

	a.queryOptions = &types.QueryOptions{
		ContractAddr: contractAddress,
		QueryMsg:     []byte{},
	}
}

func (a *bvsSquaringImpl) CreateNewTask(ctx context.Context, input int64) (*coretypes.ResultTx, error) {
	msg := types.CreateNewTaskReq{
		CreateNewTask: types.CreateNewTask{
			Input: input,
		},
	}

	msgBytes, err := json.Marshal(msg)
	(*a.executeOptions).ExecuteMsg = msgBytes

	if err != nil {
		return nil, err
	}

	return a.io.SendTransaction(ctx, *a.executeOptions)
}

func (a *bvsSquaringImpl) RespondToTask(ctx context.Context, taskId int64, result int64, operators string) (*coretypes.ResultTx, error) {
	msg := types.RespondToTaskReq{
		RespondToTask: types.RespondToTask{
			TaskId:    taskId,
			Result:    result,
			Operators: operators,
		},
	}

	msgBytes, err := json.Marshal(msg)
	(*a.executeOptions).ExecuteMsg = msgBytes

	if err != nil {
		return nil, err
	}

	return a.io.SendTransaction(ctx, *a.executeOptions)
}

func (a *bvsSquaringImpl) GetTaskInput(taskId int64) (*wasmtypes.QuerySmartContractStateResponse, error) {
	msg := types.GetTaskInputReq{
		GetTaskInput: types.GetTaskInput{
			TaskId: taskId,
		},
	}

	msgBytes, err := json.Marshal(msg)
	(*a.queryOptions).QueryMsg = msgBytes

	if err != nil {
		return nil, err
	}

	return a.io.QueryContract(*a.queryOptions)
}

func (a *bvsSquaringImpl) GetTaskResult(taskId int64) (*wasmtypes.QuerySmartContractStateResponse, error) {
	msg := types.GetTaskResultReq{
		GetTaskResult: types.GetTaskResult{
			TaskId: taskId,
		},
	}

	msgBytes, err := json.Marshal(msg)
	(*a.queryOptions).QueryMsg = msgBytes

	if err != nil {
		return nil, err
	}

	return a.io.QueryContract(*a.queryOptions)
}

func NewBVSSquaring(chainIO io.ChainIO) BVSSquaring {
	return &bvsSquaringImpl{
		io: chainIO,
	}
}
