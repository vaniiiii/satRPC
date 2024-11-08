## Aggregator Service

The aggregator service collects results from performers and attesters, manages consensus, and submits final results to the blockchain.

### API Endpoints

```plaintext
POST /api/aggregator     # Submit task results
GET /api/aggregator/task/:taskId  # Retrieve performer's data
```

### Task Submission Flow

1. **Data Collection**

```go
type Payload struct {
    TaskId    uint64 // Task identifier
    Result    string // Block data or attestation
    Timestamp int64  // Submission time
    Signature string // Signed message
    PubKey    string // Operator's public key
    Role      string // "performer" or "attester"
}
```

2. **Validation Checks**

- Signature verification
- Timestamp validity (within 2 minutes)
- Role-specific result format:
  - Performer: `blockNumber-blockHash`
  - Attesters: `true` or `false`

3. **Consensus Processing**

```plaintext
Consensus Requirements:
- Minimum attesters: configurable (default: 1)
- Consensus threshold: 66%
- Result determination:
  - Positive consensus (≥66% true): Result = 1
  - Negative consensus (≥66% false): Result = 0
```

### Task Verification Storage

```go
type TaskVerification struct {
    Performer  *TaskSubmission
    Attesters  map[string]*TaskSubmission
}

type TaskSubmission struct {
    Address   string
    Result    string
    Timestamp int64
    Role      string
}
```

### Consensus Process

1. **Collection Phase**

   - Store performer's result
   - Collect attester validations
   - Track votes for/against performer's data

2. **Vote Calculation**

   - Calculate percentage of positive/negative votes
   - Require minimum number of attesters
   - Apply consensus threshold

3. **Result Submission**
   - When consensus reached:
     - Queue final result for blockchain submission
     - Mark task as finished
     - Store result for 24 hours
