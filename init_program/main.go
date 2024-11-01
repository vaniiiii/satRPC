package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/satlayer/hello-world-bvs/init_program/core"
	"github.com/satlayer/satlayer-api/chainio/api"
	"github.com/satlayer/satlayer-api/chainio/io"
	"github.com/satlayer/satlayer-api/chainio/types"
	"github.com/satlayer/satlayer-api/logger"
	transactionprocess "github.com/satlayer/satlayer-api/metrics/indicators/transaction_process"
)

func main() {
	approverAccount, approverAddress := getApproverAccount()
	print("approverAddress: ", approverAddress)
	registerBvsContract()
	registerOperators(approverAddress)
	registerStrategy()
	registerStakers(approverAccount)
}

func getApproverAccount() (client.Account, string) {
	elkLogger := logger.NewELKLogger("bvs_demo")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	approverClient, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, core.C.Account.KeyDir, core.C.Account.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             5,
		RetryInterval:          3 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	approverClient, err = approverClient.SetupKeyring(core.C.Account.ApproverKeyName, core.C.Account.KeyringBackend)
	if err != nil {
		panic(err)
	}
	approverAccount, err := approverClient.GetCurrentAccount()
	if err != nil {
		panic(err)
	}
	approverAddress := approverAccount.GetAddress().String()

	return approverAccount, approverAddress
}

func registerBvsContract() string {
	elkLogger := logger.NewELKLogger("bvs_demo")
	elkLogger.SetLogLevel("info")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, core.C.Account.KeyDir, core.C.Account.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             5,
		RetryInterval:          3 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	chainIO, err = chainIO.SetupKeyring(core.C.Account.CallerKeyName, core.C.Account.KeyringBackend)
	if err != nil {
		panic(err)
	}
	bvsDriver := api.NewBvsDriverImpl(chainIO)
	bvsDriver.BindClient(core.C.BvsContract.BvsDriverAddr)
	txResp, err := bvsDriver.SetRegisteredBvsContract(context.Background(), core.C.BvsContract.BvsContractAddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("registerBvsContract success, txn: %s\n", txResp.Hash.String())

	stateBank := api.NewStateBankImpl(chainIO)
	stateBank.BindClient(core.C.BvsContract.BvsStateBankAddr)
	txResp, err = stateBank.SetRegisteredBvsContract(context.Background(), core.C.BvsContract.BvsContractAddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("registerBvsContract success, txn: %s\n", txResp.Hash.String())

	txResp, err = api.NewBVSDirectoryImpl(chainIO, core.C.BvsContract.BvsDirectoryAddr).RegisterBVS(context.Background(), core.C.BvsContract.BvsContractAddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("registerBvsContract success, txn: %s\n", txResp.Hash.String())
	return txResp.Hash.String()
}

func registerOperators(approverAddress string) {
	elkLogger := logger.NewELKLogger("bvs_demo")
	elkLogger.SetLogLevel("info")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, core.C.Account.KeyDir, core.C.Account.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             5,
		RetryInterval:          3 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	for _, operator := range core.C.Account.OperatorsKeyName {
		chainIO, err = chainIO.SetupKeyring(operator, core.C.Account.KeyringBackend)
		if err != nil {
			panic(err)
		}
		account, err := chainIO.GetCurrentAccount()
		if err != nil {
			panic(err)
		}

		delegation := api.NewDelegationImpl(chainIO, core.C.BvsContract.DelegatorAddr)
		txResp, err := delegation.RegisterAsOperator(
			context.Background(),
			account.GetPubKey(),
			"",
			approverAddress,
			"",
			0,
		)
		if err != nil {
			fmt.Println("Ere registerAsOperator to delegation failed: ", err)
		} else {
			fmt.Println("registerAsOperator to delegation success:", txResp)
		}
		// register operator to bvsDirectory
		txResp, err = api.NewBVSDirectoryImpl(chainIO, core.C.BvsContract.BvsDirectoryAddr).RegisterOperator(context.Background(), account.GetAddress().String(), account.GetPubKey())
		if err != nil {
			fmt.Println("Err: registerOperators to bvsDirectory failed: ", err)
		} else {
			fmt.Println("registerOperators to bvsDirectory success:", txResp)
		}

	}
}

func registerStrategy() {
	elkLogger := logger.NewELKLogger("bvs_demo")
	elkLogger.SetLogLevel("info")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, core.C.Account.KeyDir, core.C.Account.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             5,
		RetryInterval:          3 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	chainIO, err = chainIO.SetupKeyring(core.C.Account.CallerKeyName, core.C.Account.KeyringBackend)
	strategyManager := api.NewStrategyManager(chainIO)
	strategyManager.BindClient(core.C.BvsContract.StrategyMangerAddr)
	ctx := context.Background()

	// register delegation manager
	resp, err := strategyManager.SetDelegationManager(ctx, core.C.BvsContract.DelegatorAddr)
	if err != nil {
		fmt.Println("Err: setDelegationManager failed: ", err)
	} else {
		fmt.Println("SetDelegationManager success:", resp)
	}

	resp, err = strategyManager.AddStrategiesToWhitelist(ctx, []string{core.C.BvsContract.StrategyAddr}, []bool{false})
	if err != nil {
		fmt.Println("Err: addStrategiesToWhitelist failed: ", err)
	} else {
		fmt.Println("AddStrategiesToWhitelist success:", resp)
	}
}

func registerStakers(approverAccount client.Account) {
	elkLogger := logger.NewELKLogger("bvs_demo")
	elkLogger.SetLogLevel("info")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, core.C.Account.KeyDir, core.C.Account.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             5,
		RetryInterval:          3 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	for _, staker := range core.C.StakerOperatorMap {
		fmt.Printf("staker: %+v\n", staker)
		sclient, err := chainIO.SetupKeyring(staker.StakerKeyName, core.C.Account.KeyringBackend)
		if err != nil {
			panic(err)
		}
		delegation := api.NewDelegationImpl(sclient, core.C.BvsContract.DelegatorAddr)
		oClient, err := chainIO.SetupKeyring(staker.OperatorKeyName, core.C.Account.KeyringBackend)
		if err != nil {
			panic(err)
		}
		acc, err := oClient.GetCurrentAccount()
		if err != nil {
			panic(err)
		}
		txResp, err := delegation.DelegateTo(
			context.Background(),
			acc.GetAddress().String(),
			approverAccount.GetAddress().String(),
			core.C.Account.ApproverKeyName,
			approverAccount.GetPubKey(),
		)
		if err != nil {
			fmt.Println("Err: ", err)
		}
		fmt.Println("DelegateTo to operator success:", txResp)

		txnResp, err := api.IncreaseTokenAllowance(context.Background(), sclient, 9999999999999999, core.C.BvsContract.Cw20TokenAddr, core.C.BvsContract.StrategyMangerAddr, sdktypes.NewInt64DecCoin("ubbn", 1))
		if err != nil {
			fmt.Println("Err: ", err)
		}
		fmt.Println("increaseTokenAllowance success:", txnResp)

		// register staker to strategy
		strategyManager := api.NewStrategyManager(sclient)
		strategyManager.BindClient(core.C.BvsContract.StrategyMangerAddr)
		resp, err := strategyManager.DepositIntoStrategy(context.Background(), core.C.BvsContract.StrategyAddr, core.C.BvsContract.Cw20TokenAddr, staker.Amount)
		if err != nil {
			err := fmt.Errorf("DepositIntoStrategy failed: %v", err)
			fmt.Println("Err", err)
		} else {
			fmt.Println("DepositIntoStrategy success:", resp)
		}

	}
}

func approve() {
	elkLogger := logger.NewELKLogger("bvs_demo")
	elkLogger.SetLogLevel("info")
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	chainIO, err := io.NewChainIO(core.C.Chain.Id, core.C.Chain.Rpc, core.C.Account.KeyDir, core.C.Account.Bech32Prefix, elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             3,
		RetryInterval:          1 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		panic(err)
	}
	sclient, err := chainIO.SetupKeyring("uploader", core.C.Account.KeyringBackend)
	txnResp, err := api.IncreaseTokenAllowance(context.Background(), sclient, 9999999999999999, core.C.BvsContract.Cw20TokenAddr, core.C.BvsContract.RewardCoordinatorAddr, sdktypes.NewInt64DecCoin("ubbn", 1))
	if err != nil {
		fmt.Println("Err: ", err)
	}
	fmt.Println("increaseTokenAllowance success:", txnResp)

}
