package node

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	rio "io"
	"io/ioutil"
	"net/http"
	"strconv"
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
		return
	}
	task, err := strconv.Atoi(taskId)
	if err != nil {
		fmt.Println("format err:", err)
		return
	}
	_, address, err := util.PubKeyToAddress(n.pubKeyStr)
	if err != nil {
		panic(err)
	}

	// Get the latest block data regardless of role
	latestBlockNumber, latestBlockHash, err := n.fetchLatestBlockData()
	if err != nil {
		panic(err)
	}

	fmt.Println("Latest Block Number: ", latestBlockNumber)
	fmt.Println("Latest Block Hash: ", latestBlockHash)

	// Determine role based on selected performer
	role := "attester"
	if value == address {
		role = "performer"
	}

	// Send data to aggregator with appropriate role
	err = n.sendAggregator(int64(task), latestBlockNumber, latestBlockHash, role)
	if err != nil {
		panic(err)
	}
	return
}

// fetchLatestBlockData retrieves the latest block number and its hash from the configured RPC endpoint.
//
// Returns the block number as an int64 and the block hash as a string.
// If there is an error during the retrieval, an error will be returned.
func (n *Node) fetchLatestBlockData() (int64, string, error) {
	// JSON-RPC payload to get the latest block number
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	// Send POST request to the RPC endpoint
	resp, err := http.Post(core.C.Rpc.Endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return 0, "", fmt.Errorf("failed to send request to RPC endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Extract the block number from the response
	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, "", fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Convert the block number from hex string to int64
	var latestBlockNumber int64
	if _, err := fmt.Sscanf(result.Result, "0x%x", &latestBlockNumber); err != nil {
		return 0, "", fmt.Errorf("failed to parse block number: %v", err)
	}

	// Prepare payload to fetch the block hash by block number
	payload = map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{result.Result, false},
		"id":      1,
	}
	data, err = json.Marshal(payload)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal JSON payload for block hash: %v", err)
	}

	// Send POST request to get the block details
	resp, err = http.Post(core.C.Rpc.Endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return 0, "", fmt.Errorf("failed to send request for block hash: %v", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response body for block hash: %v", err)
	}

	// Extract the block hash from the response
	var blockResult struct {
		Result struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &blockResult); err != nil {
		return 0, "", fmt.Errorf("failed to parse block hash response: %v", err)
	}

	return latestBlockNumber, blockResult.Result.Hash, nil
}

// sendAggregator sends the task result to the aggregator.
//
// taskId is the unique identifier of the task.
// latestBlockNumber is the number of the latest block in the chain.
// latestBlockHash is the hash of the latest block in the chain.
// Returns an error if there is an issue with the sending process.
func (n *Node) sendAggregator(taskId int64, latestBlockNumber int64, latestBlockHash string, role string) (err error) {
	nowTs := time.Now().Unix()
	result := fmt.Sprintf("%d-%s", latestBlockNumber, latestBlockHash)
	msgPayload := fmt.Sprintf("%s-%d-%d-%s", core.C.Chain.BvsHash, nowTs, taskId, result)
	core.L.Info(fmt.Sprintf("msgPayload: %s\n", msgPayload))
	signature, err := n.chainIO.GetSigner().Sign([]byte(msgPayload))

	payload := Payload{
		TaskId:    taskId,
		Result:    result,
		Timestamp: nowTs,
		Signature: signature,
		PubKey:    n.pubKeyStr,
		Role:      role,
	}
	fmt.Printf("task result send aggregator payload: %+v\n", payload)
	if err != nil {
		return
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %s", err)
		return
	}

	resp, err := http.Post(core.C.Aggregator.Url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending aggregator : %s\n", err)
		return
	}
	if resp.StatusCode != 200 {
		body, _ := rio.ReadAll(resp.Body)
		fmt.Printf("Error sending aggregator : %s\n", string(body))
		return
	}
	return
}
