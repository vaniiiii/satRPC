package tests

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/satlayer/satlayer-api/chainio/types"
	"github.com/satlayer/satlayer-api/logger"
	transactionprocess "github.com/satlayer/satlayer-api/metrics/indicators/transaction_process"

	"github.com/satlayer/satlayer-api/chainio/io"

	"github.com/satlayer/satlayer-api/signer"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestSign is a test function that tests the functionality of the signer.
//
// It initializes a CosmosClient with the provided chainID, rpcURI,
// homeDir, keyName, and chainName. It then signs a message using the
// initialized signer and verifies the signature.
//
// Parameters:
//   - t: *testing.T - the testing object used for logging and reporting test
//     failures.
//
// Return type: None.
func TestSign(t *testing.T) {
	msg := "hello world"
	chainID := "sat-bbn-testnet1"
	rpcURI := "https://rpc.sat-bbn-testnet1.satlayer.net"
	homeDir := "../../.babylond" // Please refer to the readme to obtain
	keyName := "operator1"       // Please refer to the readme to obtain
	elkLogger := logger.NewMockELKLogger()
	reg := prometheus.NewRegistry()
	metricsIndicators := transactionprocess.NewPromIndicators(reg, "bvs_demo")
	cs, err := io.NewChainIO(chainID, rpcURI, homeDir, "bbn", elkLogger, metricsIndicators, types.TxManagerParams{
		MaxRetries:             3,
		RetryInterval:          1 * time.Second,
		ConfirmationTimeout:    60 * time.Second,
		GasPriceAdjustmentRate: "1.1",
	})
	if err != nil {
		t.Fatalf("failed to create chain IO: %s", err)
		return
	}
	cs, err = cs.SetupKeyring(keyName, "test")
	if err != nil {
		t.Fatalf("failed to setup keyring: %s", err)
		return
	}
	msgByte := []byte(msg)
	signature, err := cs.GetSigner().Sign(msgByte)
	if err != nil {
		t.Fatalf("failed to sign: %v\n", err)
		return
	}
	t.Logf("%+v\n", signature)

	account, err := cs.GetCurrentAccount()
	if err != nil {
		t.Fatalf("failed to get account: %v\n", err)
		return
	}
	t.Logf("%+v\n", account)
	pubKey := account.GetPubKey()
	pubKeyBytes := pubKey.Bytes()
	pubKeyStr := base64.StdEncoding.EncodeToString(pubKeyBytes)
	t.Logf("pubKeyStr: %s\n", pubKeyStr)

	pubKeyRawBytes, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		t.Fatalf("failed to decode public key: %v\n", err)
		return
	}

	newPubKey := secp256k1.PubKey{Key: pubKeyRawBytes}

	address := sdk.AccAddress(newPubKey.Address()).String()
	t.Logf("address: %s\n", address)
	verifyResult, err := signer.VerifySignature(&newPubKey, msgByte, signature)
	if err != nil {
		t.Fatalf("failed to verify signature: %v\n", err)
		return
	}
	t.Logf("%+v\n", verifyResult)

}
