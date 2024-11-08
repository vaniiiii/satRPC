## BVS Operator Logic

The BVS operator logic implements the off-chain validation system for satRPC, with operators serving in two distinct roles: performer and attester.

### Task Processing Flow

1. Operator monitors the BVS Driver contract for new tasks
2. Upon task creation, checks if selected as performer
3. Processes task based on assigned role
4. Submits results to aggregator

### Performer Role

```go
// When operator is selected as performer
if value == address {
    // Fetch latest block data
    latestBlockNumber, latestBlockHash := fetchLatestBlockData()

    // Format result as "blockNumber-blockHash"
    result := fmt.Sprintf("%d-%s", latestBlockNumber, latestBlockHash)

    // Send to aggregator
    sendAggregator(taskId, result, "performer")
}
```

The performer:

- Fetches the latest block number and hash from their RPC endpoint
- Formats and signs the data
- Submits data to the aggregator node

### Attester Role

```go
// When operator is an attester
// Retrieves performer's submitted data
performerData := getPerformerData(taskId)

// Validates:
// 1. Block exists and hash matches
// 2. Block is recent (within 10 blocks)
// 3. Correct performer submitted the data
isValid := validatePerformerData(performerData)

// Submit attestation
sendAggregator(taskId, isValid, "attester")
```

Attesters:

- Retrieve performer's submitted block data
- Validate the block information
- Submit true/false attestation to aggregator

### Validation Criteria

Attesters verify three key aspects:

1. Block Authenticity: Hash matches the reported block number
2. Timeliness: Block is within 10 blocks of current height
3. Authorization: Data was submitted by the designated performer

### Data Submission

Both roles submit signed payloads to the aggregator:

```go
type Payload struct {
    TaskId    int64  // Task identifier
    Result    string // Block data or attestation result
    Timestamp int64  // Submission timestamp
    Signature string // Signed message
    PubKey    string // Operator's public key
    Role      string // "performer" or "attester"
}
```

### Error Handling

- Performers: Retry logic for RPC endpoint failures
- Attesters: Multiple attempts to fetch performer data
- Both: Signature verification and validation checks
