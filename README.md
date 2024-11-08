# 🔐 satRPC

## Introduction 🌐

Welcome to the satRPC project! Using Babylon and SatLayer Stack, we're building a decentralized RPC network that helps strengthen blockchain infrastructure. By leveraging Bitcoin restaking through SatLayer and Babylon, we're bringing Bitcoin's robust security model to RPC services, making them more decentralized and trustless. Our network provides reliable access to blockchain data while maintaining the security guarantees that Bitcoin's offers. 🌟

## How It Works 🔍

### Network Operators 🌟

Operators are at the core of our network, playing a vital role in maintaining the decentralized RPC infrastructure. Their responsibilities include:

1. **Node Operation**: Running a Babylon node or a node of any network they wish to provide RPC endpoints for
2. **BVS Logic**: Running the Bitcoin Validated Service logic to ensure network integrity
3. **RPC Service**: Providing secure and reliable RPC endpoints to users
4. **Network Validation**: Participating in the network's validation process to verify other operators' results

### Process Overview 🛠️

The decentralized RPC process follows a structured approach to ensure data accuracy and network reliability:

1. **Task Initiation**: Every 10 blocks, a new task is created for selecting a specific operator to be the performer
2. **Performer Action**: The selected operator's BVS logic:
   - Verifies their address matches the task
   - Fetches the latest block from their RPC endpoint
   - Broadcasts the block data and hash to the network
3. **Network Attestation**: Other operators validate the performer's data by:
   - Verifying the provided block information
   - Submitting true/false attestations to the aggregator
4. **Result Aggregation**: The aggregator node:
   - Collects all operator attestations
   - Determines consensus (requires 66% agreement)
   - Submits the final result to the blockchain
5. **Reputation Update**: Based on the aggregated result:
   - Successful performers gain reputation points
   - Failed or incorrect submissions lose reputation points

## Architecture Overview 🏗️

Built using Babylon, SatLayer, CosmWasm, CosmJs, and Vite 🛠️

### Components 🧩

- **Babylon Node**: Core infrastructure required for network operation and RPC endpoint provision
- **BVS Logic**: Software that operators run to:
  - Listen for new tasks
  - Perform block data fetching when selected
  - Validate other operators' results
- **Aggregator Node**: Central component that:
  - Collects attestations from operators
  - Determines consensus
  - Submits final results to blockchain
- **satRPC Smart Contract**: CosmWasm-based contract that:

  - Tracks and updates operator reputation scores

### Flow Diagram 🔄

1. **Network Entry**: Operator sets up Babylon node and BVS logic software
2. **Task Creation**: Every 10 blocks new task is created
3. **Performer Selection**: Task specifies which operator will fetch and broadcast block data
4. **Task Execution**: Selected operator retrieves and broadcasts latest block information
5. **Attestation**: Other operators validate and attest to the performer's data
6. **Consensus Building**: Aggregator collects attestations until 66% threshold is reached
7. **Score Update**: Smart contract updates operator's reputation based on consensus result

## Conclusion 🌟

satRPC demonstrates the potential of Babylon and SatLayer's restaking infrastructure as powerful building blocks for creating decentralized services. By leveraging Bitcoin's security through restaking, we're not just creating a decentralized RPC network - we're showcasing how Babylon Validated Services (BVS) can support diverse distributed validation systems.

The same architecture we use for satRPC can be applied to build various systems requiring distributed validation semantics.

---

Feel free to reach out if you have any questions or want to contribute! 🌐💬

## Requirements

[Coming Soon]
