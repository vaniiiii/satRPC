package node

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	rio "io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/satlayer/satlayer-api/chainio/types"
	"github.com/satlayer/satlayer-api/logger"
	transactionprocess "github.com/satlayer/satlayer-api/metrics/indicators/transaction_process"

	"github.com/satlayer/satlayer-api/chainio/io"

	"github.com/satlayer/hello-world-bvs/aggregator/util"
	"github.com/satlayer/hello-world-bvs/bvs_offchain/core"
	"github.com/satlayer/satlayer-api/chainio/api"
	"github.com/satlayer/satlayer-api/chainio/indexer"
)

type Node struct {
	bvsContract string
	pubKeyStr   string
	chainIO     io.ChainIO
	stateBank   api.StateBank
}

type Payload struct {
	TaskId    int64  `json:"taskID"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
	PubKey    string `json:"pubKey"`
	Role      string `json:"role"`
}

type PerformerData struct {
	Result  string `json:"result"`
	Address string `json:"address"`
}

// NewNode creates a new Node instance with the given configuration.
//
// It initializes a new Cosmos client, retrieves the account, and sets up the BVS contracts and state bank.
// Returns a pointer to the newly created Node instance.
// NewNode creates a new Node instance with the given configuration.
//
// It initializes a new Cosmos client, retrieves the account, and sets up the BVS contracts and state bank.
// Returns a pointer to the newly created Node instance.
func NewNode() *Node {
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
	chainIO, err = chainIO.SetupKeyring(core.C.Owner.KeyName, core.C.Owner.KeyringBackend)
	if err != nil {
		panic(err)
	}
	account, err := chainIO.GetCurrentAccount()
	if err != nil {
		panic(err)
	}
	pubKeyStr := base64.StdEncoding.EncodeToString(account.GetPubKey().Bytes())
	txResp, err := api.NewBVSDirectoryImpl(chainIO, core.C.Chain.BvsDirectory).GetBVSInfo(core.C.Chain.BvsHash)
	if err != nil {
		panic(err)
	}
	stateBank := api.NewStateBankImpl(chainIO)

	return &Node{
		bvsContract: txResp.BVSContract,
		stateBank:   stateBank,
		chainIO:     chainIO,
		pubKeyStr:   pubKeyStr,
	}
}

// Run starts the node's main execution loop.
//
// ctx is the context for the Run function.
// No return value.
// Run starts the node's main execution loop.
//
// ctx is the context for the Run function.
// No return value.
func (n *Node) Run(ctx context.Context) {
	// @reminder ask what this is doing?
	if err := n.syncStateBank(ctx); err != nil {
		panic(err)
	}
	if err := n.monitorDriver(ctx); err != nil {
		panic(err)
	}
}

// syncStateBank synchronizes the state bank with the latest blockchain state.
//
// ctx is the context for the syncStateBank function.
// Returns an error if the synchronization fails.
// syncStateBank synchronizes the state bank with the latest blockchain state.
//
// ctx is the context for the syncStateBank function.
// Returns an error if the synchronization fails.
func (n *Node) syncStateBank(ctx context.Context) (err error) {
	res, err := n.chainIO.QueryNodeStatus(ctx)
	if err != nil {
		panic(err)
	}
	latestBlock := res.SyncInfo.LatestBlockHeight
	idx := n.stateBank.Indexer(n.chainIO.GetClientCtx(), core.C.Chain.StateBank, n.bvsContract, latestBlock, []string{"wasm-UpdateState"}, 1, 10)
	processingQueue, err := idx.Run(ctx)
	if err != nil {
		panic(err)
	}
	go func() {
		n.stateBank.EventHandler(processingQueue)
	}()
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if idx.IsUpToDate {
				return
			}
		}
	}

	return
}

// monitorDriver monitors the driver contract for events and performs actions based on the event type.
//
// ctx is the context for the monitorDriver function.
// Returns an error if there is an issue with the monitoring process.
// monitorDriver monitors the driver contract for events and performs actions based on the event type.
//
// ctx is the context for the monitorDriver function.
// Returns an error if there is an issue with the monitoring process.
func (n *Node) monitorDriver(ctx context.Context) (err error) {
	res, err := n.chainIO.QueryNodeStatus(ctx)
	if err != nil {
		panic(err)
	}
	latestBlock := res.SyncInfo.LatestBlockHeight
	fmt.Println("latestBlock: ", latestBlock)
	evtIndexer := indexer.NewEventIndexer(
		n.chainIO.GetClientCtx(),
		core.C.Chain.BvsDriver,
		latestBlock,
		[]string{"wasm-ExecuteBVSOffchain"},
		1,
		10)
	evtChain, err := evtIndexer.Run(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("chain: ", evtChain)
	for evt := range evtChain {
		if evt.AttrMap["sender"] != n.bvsContract {
			continue
		}
		switch evt.EventType {
		case "wasm-ExecuteBVSOffchain":
			time.Sleep(5 * time.Second)
			taskId := evt.AttrMap["task_id"]
			fmt.Println("taskId: ", taskId)
			if err := n.calcTask(taskId); err != nil {
				fmt.Println("ExecuteBVSOffchain error: ", err)
			}
		default:
			fmt.Println("unhandled event: ", evt.EventType)
		}
	}
	return
}

// calcTask calculates the task result and sends it to the aggregator.
//
// taskId is the unique identifier of the task.
// Returns an error if there is an issue with the calculation or sending process.
// calcTask calculates the task result and sends it to the aggregator.
//
// taskId is the unique identifier of the task.
// Returns an error if there is an issue with the calculation or sending process.
func (n *Node) calcTask(taskId string) (err error) {
	stateKey := fmt.Sprintf("taskId.%s", taskId)
	value, err := n.stateBank.GetWasmUpdateState(stateKey)
	if err != nil {
		return err
	}

	task, err := strconv.Atoi(taskId)
	if err != nil {
		fmt.Println("format err:", err)
		return err
	}
	_, address, err := util.PubKeyToAddress(n.pubKeyStr)
	if err != nil {
		panic(err)
	}

	// Check if we're the performer
	if value == address {
		fmt.Printf("Selected as performer for task %s\n", taskId)
		latestBlockNumber, latestBlockHash, err := n.fetchLatestBlockData()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Performer data - Block Number: %d, Hash: %s\n", latestBlockNumber, latestBlockHash)

		result := fmt.Sprintf("%d-%s", latestBlockNumber, latestBlockHash)
		if err = n.sendAggregator(int64(task), result, "performer"); err != nil {
			panic(err)
		}
		return nil
	}

	// We're an attester, try multiple times to get performer's data
	fmt.Printf("Acting as attester for task %s, waiting for performer data...\n", taskId)

	// Retry configuration
	maxRetries := 5               // Try 5 times
	retryDelay := 3 * time.Second // Wait 3 seconds between tries

	var performerData PerformerData
	var gotData bool

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("Retry %d/%d for task %s\n", i+1, maxRetries, taskId)
		}

		// Try to get performer's data
		url := fmt.Sprintf("%s/task/%s", core.C.Aggregator.Url, taskId)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to get performer data: %v, retrying...\n", err)
			time.Sleep(retryDelay)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			fmt.Printf("No performer data yet for task %s, retrying...\n", taskId)
			time.Sleep(retryDelay)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&performerData); err != nil {
			resp.Body.Close()
			fmt.Printf("Failed to decode performer data: %v, retrying...\n", err)
			time.Sleep(retryDelay)
			continue
		}
		resp.Body.Close()

		gotData = true
		fmt.Printf("Got performer data for task %s: %s\n", taskId, performerData.Result)
		break
	}

	if !gotData {
		return fmt.Errorf("failed to get performer data after %d retries", maxRetries)
	}

	isValid, err := n.validatePerformerData(performerData, value)
	if err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	result := "true"
	if !isValid {
		result = "false"
	}

	if err = n.sendAggregator(int64(task), result, "attester"); err != nil {
		panic(err)
	}

	fmt.Printf("Successfully sent attestation for task %s\n", taskId)
	return nil
}

// fetchLatestBlockData retrieves the latest block number and its hash from the configured Babylon RPC endpoint.
//
// Returns the block number as an int64 and the block hash as a string.
// If there is an error during the retrieval, an error will be returned.
func (n *Node) fetchLatestBlockData() (int64, string, error) {
	// First, get the status to find the latest block height
	resp, err := http.Get(fmt.Sprintf("%s/status", core.C.Rpc.Endpoint))
	if err != nil {
		return 0, "", fmt.Errorf("failed to get status: %v", err)
	}
	defer resp.Body.Close()

	var statusResp core.StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return 0, "", fmt.Errorf("failed to decode status response: %v", err)
	}

	// Convert block height from string to int64
	latestHeight, err := strconv.ParseInt(statusResp.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse block height: %v", err)
	}

	// Now fetch the block details using the height
	blockResp, err := http.Get(fmt.Sprintf("%s/block?height=%d", core.C.Rpc.Endpoint, latestHeight))
	if err != nil {
		return 0, "", fmt.Errorf("failed to get block: %v", err)
	}
	defer blockResp.Body.Close()

	var block core.BlockResponse
	if err := json.NewDecoder(blockResp.Body).Decode(&block); err != nil {
		return 0, "", fmt.Errorf("failed to decode block response: %v", err)
	}

	return latestHeight, block.Result.BlockID.Hash, nil
}

func (n *Node) validatePerformerData(performerData PerformerData, expectedAddress string) (bool, error) {
	// Parse performer's data
	parts := strings.Split(performerData.Result, "-")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid performer data format")
	}

	performerBlockNumber, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse block number: %v", err)
	}
	performerBlockHash := parts[1]

	// Get current block data for verification
	currentBlockNumber, _, err := n.fetchLatestBlockData()
	if err != nil {
		return false, fmt.Errorf("failed to fetch current block data: %v", err)
	}

	// Verify block exists and hash matches
	resp, err := http.Get(fmt.Sprintf("%s/block?height=%d", core.C.Rpc.Endpoint, performerBlockNumber))
	if err != nil {
		return false, fmt.Errorf("failed to fetch block data: %v", err)
	}
	defer resp.Body.Close()

	var blockResponse core.BlockResponse
	if err := json.NewDecoder(resp.Body).Decode(&blockResponse); err != nil {
		return false, fmt.Errorf("failed to decode block data: %v", err)
	}

	// Get the hash from the block response
	actualBlockHash := blockResponse.Result.BlockID.Hash

	isBlockValid := actualBlockHash == performerBlockHash
	isBlockRecent := currentBlockNumber-performerBlockNumber <= 10 // You might want to adjust this threshold
	isCorrectPerformer := performerData.Address == expectedAddress

	return isBlockValid && isBlockRecent && isCorrectPerformer, nil
}

// sendAggregator sends the task result to the aggregator.
//
// taskId is the unique identifier of the task.
// result is either block data (for performer) or validation result (for attester)
// role is either "performer" or "attester"
// Returns an error if there is an issue with the sending process.
func (n *Node) sendAggregator(taskId int64, result string, role string) (err error) {
	nowTs := time.Now().Unix()

	// Create message payload based on role
	msgPayload := fmt.Sprintf("%s-%d-%d-%s", core.C.Chain.BvsHash, nowTs, taskId, result)
	core.L.Info(fmt.Sprintf("msgPayload: %s\n", msgPayload))

	signature, err := n.chainIO.GetSigner().Sign([]byte(msgPayload))
	if err != nil {
		return fmt.Errorf("failed to sign payload: %v", err)
	}

	payload := Payload{
		TaskId:    taskId,
		Result:    result, // For performer: "blockNum-hash", for attester: "true"/"false"
		Timestamp: nowTs,
		Signature: signature,
		PubKey:    n.pubKeyStr,
		Role:      role,
	}

	fmt.Printf("Sending to aggregator - Role: %s, TaskId: %d, Result: %s\n", role, taskId, result)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %s\n", err)
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(core.C.Aggregator.Url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending to aggregator: %s\n", err)
		return fmt.Errorf("failed to send to aggregator: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := rio.ReadAll(resp.Body)
		fmt.Printf("Error from aggregator: %s\n", string(body))
		return fmt.Errorf("aggregator returned non-200 status: %d, body: %s", resp.StatusCode, body)
	}

	fmt.Printf("Successfully sent %s data to aggregator for task %d\n", role, taskId)
	return nil
}
