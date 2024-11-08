## Smart Contract

The satRPC smart contract is built using CosmWasm and manages task creation, operator performance tracking, and score maintenance for the RPC network.

### Core Components

- **Task Creation**: Creates tasks every 10 blocks with a designated performer
- **Score Management**: Tracks operator performance through a scoring system
- **State Storage**: Maintains task history and operator statistics

### Task Creation Flow

```rust
CreateNewTask {
    input: Addr  // Operator address selected to perform the task
}
```

1. Each task is assigned a unique incremental ID
2. Task stores the selected operator's address
3. Contract emits two messages:
   - To State Bank: Stores task-operator mapping
   - To BVS Driver: Triggers off-chain BVS logic execution

### Score Management

```rust
RespondToTask {
    task_id: u64,  // Task identifier
    result: i64    // 1 for success, 0 for failure
}
```

- Only the designated aggregator can submit results
- Operator scores are updated based on performance:
  - Success (1): Score increases by 1
  - Failure (0): Score decreases by 1
- Each task increments the operator's max score counter

### Key State Variables

- `CREATED_TASKS`: Maps task IDs to assigned operators
- `OPERATOR_SCORE`: Current performance score of each operator
- `OPERATOR_MAX_SCORE`: Total tasks assigned to each operator
- `RESPONDED_TASKS`: Stores task results submitted by aggregator

### Query Functions

```rust
GetOperatorScore { operator: Addr }    // Current operator score
GetOperatorMaxScore { operator: Addr } // Total tasks assigned
GetTaskInput { task_id: u64 }         // Task performer
GetTaskResult { task_id: u64 }        // Task result
```
