# 🔐 satRPC

## Contents

- [Introduction](#introduction)
- [How It Works](#how-it-works)
  - [Network Operators](#network-operators)
  - [Process Overview](#process-overview)
- [Architecture Overview](#architecture-overview)
  - [Components](#components)
  - [Flow Diagram](#flow-diagram)
- [Technical Documentation](#documentation)
- [Running the BVS](#running-the-bvs)
  - [Prerequisites](#prerequisites)
  - [Setting up Wallets](#setting-up-wallets)
  - [Build BVS Contract](#build-bvs-contract)
  - [Register Operators](#register-operators)
  - [Configure Environment Files](#configure-environment-files)
  - [Start the System](#start-the-system)
- [Conclusion](#conclusion)

## Introduction

Welcome to the satRPC project! Using Babylon and SatLayer Stack, we're building a decentralized RPC network that helps strengthen blockchain infrastructure. By leveraging Bitcoin restaking through SatLayer and Babylon, we're bringing Bitcoin's robust security model to RPC services, making them more decentralized and trustless. Our network provides reliable access to blockchain data while maintaining the security guarantees that Bitcoin's offers. 🌟

## How It Works

### Network Operators

Operators are at the core of our network, playing a vital role in maintaining the decentralized RPC infrastructure. Their responsibilities include:

1. **Node Operation**: Running a Babylon node or a node of any network they wish to provide RPC endpoints for
2. **BVS Logic**: Running the Bitcoin Validated Service logic to ensure network integrity
3. **RPC Service**: Providing secure and reliable RPC endpoints to users
4. **Network Validation**: Participating in the network's validation process to verify other operators' results

### Process Overview

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

## Architecture Overview

Built using Babylon, SatLayer, CosmWasm, CosmJs, and Vite 🛠️

### Components

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

### Flow Diagram

1. **Network Entry**: Operator sets up Babylon node and BVS logic software
2. **Task Creation**: Every 10 blocks new task is created
3. **Performer Selection**: Task specifies which operator will fetch and broadcast block data
4. **Task Execution**: Selected operator retrieves and broadcasts latest block information
5. **Attestation**: Other operators validate and attest to the performer's data
6. **Consensus Building**: Aggregator collects attestations until 66% threshold is reached
7. **Score Update**: Smart contract updates operator's reputation based on consensus result

[Your existing introduction content]

## Documentation

For detailed technical documentation of each component, see:

- [Smart Contract Implementation](./docs/smart-contract.md)
- [BVS Operator Guide](./docs/operator.md)
- [Aggregator Service](./docs/aggregator.md)

## Running the BVS

To run the BVS, you need to ensure that all the necessary tools and dependencies are installed on your machine. Below are the prerequisites you'll need to get started.

### Prerequisites

Before running the BVS, make sure you have the following installed:

- **Rust**: Ensure you’ve installed [Rust](https://rustup.rs/). Rust is used to compile the necessary components for running BVS logic.
- **Go**: Ensure you've installed [Go](https://go.dev)
- **Redis**: Ensure you've installed [Redis](https://redis.io/docs/latest/operate/oss_and_stack/install/install-redis/)
- **Docker**: Docker is essential for compiling the CosmWasm contract. Install Docker by following the [official installation guide](https://docs.docker.com/get-docker/).
- **CosmWasm**: Read the [CosmWasm Documentation](https://docs.cosmwasm.com/) to set up the environment for compiling and running CosmWasm-based smart contracts.
- **Babylond**: The Babylon Node is required for connecting to the satRPC network. Ensure you’ve installed the `babylond` command-line tool by following the [Babylon guide](https://docs.babylonchain.io).
- **SatLayer CLI**: To interact with SatLayer's network, you need to have the `satlayer-cli` tool installed. Follow the [satlayer-cli guide](https://docs.satlayer.xyz/~/changes/z0grj9ZdLyT0QOOYHm1R) to set it up on your machine.

### Setting up Wallets

At a minimum, you'll need four wallets, which we will name: `bvs-owner`, `bvs-operator`, `bvs-aggregator`. The easiest way to create these wallets is through the Babylon command. Please refer to our Babylon command guide for installation details.

To create the wallets, run the following commands in a terminal:

```bash
babylond keys add bvs-owner
babylond keys add bvs-operator
babylond keys add bvs-operator-2
babylond keys add bvs-user
babylond keys add bvs-aggregator
```

### Sending Tokens to Wallets (Last updated: October 30, 2024)

At time of writing, October 30, 2024, each account needs to have at least one outgoing transaction for the satlayer-cli command to verify on-chain details. It is sufficient to send 1ubbn between the accounts.

### Build BVS Contract

Here are the steps to clone the repository, compile the contract, and run tests:

#### Step 1: Clone the Repository

```
git clone https://github.com/vaniiiii/satRPC
cd contract/bvs_squaring
```

#### Step 2: Compile the Contract:

```
docker run --rm -v "$(pwd)":/code --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry cosmwasm/optimizer:0.16.0
```

#### Step 3: Deploy the Contract:

```bash
babylond tx wasm store artifacts/bvs_squaring.wasm --from=bvs-owner --gas=auto --gas-prices=1ubbn --gas-adjustment=1.3 --chain-id=sat-bbn-testnet1 -b=sync --yes --log_format=json --node https://rpc.sat-bbn-testnet1.satlayer.net
```

#### Step 4: Obtain code_id

```
curl -s https://lcd3.sat-bbn-testnet1.satlayer.net/cosmos/tx/v1beta1/txs/${TX_HASH} | jq '.tx_response.events.[] | select(.type == "store_code").attributes.[] | select(.key == "code_id").value' | sed 's/"//g'
```

#### Step 5: Instantiate the contract

```bash
./babylond tx wasm instantiate ${code_id} '{"aggregator": "<address of bvs-aggregator>", "state_bank": "bbn1h9zjs2zr2xvnpngm9ck8ja7lz2qdt5mcw55ud7wkteycvn7aa4pqpghx2q", "bvs_driver": "bbn18x5lx5dda7896u074329fjk4sflpr65s036gva65m4phavsvs3rqk5e59c"}' --from={your_wallet} --no-admin --label="bvs" --gas=auto --gas-prices=1ubbn --gas-adjustment=1.3 --chain-id=sat-bbn-testnet1 -b=sync --yes --log_format=json --node https://rpc.sat-bbn-testnet1.satlayer.net
```

#### Step 6: Get contract address

The contract address can be acquired by an API call to the node.

```bash
curl -s -X GET "https://lcd1.sat-bbn-testnet1.satlayer.net/cosmwasm/wasm/v1/code/{code_id}/contracts" -H "accept: application/json" | jq -r '.contracts[-1]'
```

#### Step 7: Register the BVS contract to BVS directory

```bash
./satlayer-cli directory reg-bvs bvs-owner ${contract_address}
```

This command will output a bvsHash - make sure to save it as you'll need it for configuring the BVS operator and aggregator.

### Register Operators

Each operator needs to be registered in both the DelegationManager and BVSDirectory. Run these commands for each operator (e.g., bvs-operator and bvs-operator-2):

```bash
# For first operator
./satlayer-cli delegation reg-operator bvs-operator [bvs-approver]
./satlayer-cli directory reg-operator bvs-operator

# For second operator
./satlayer-cli delegation reg-operator bvs-operator-2 [bvs-approver]
./satlayer-cli directory reg-operator bvs-operator-2
```

Note: `bvs-approver` is an optional address. If you don't need an approver, you can set it to 0.

### Configure Environment Files

Update the following files with your `bvsHash` from the deployment step and appropriate key names:

```toml
# aggregator/env.toml
[chain]
bvsHash = "<your-bvs-hash>"

[owner]
keyringBackend = "os"
keyName = "bvs-aggregator"

# bvs_offchain/env.toml
[chain]
bvsHash = "<your-bvs-hash>"

[owner]
keyringBackend = "os"
keyName = "bvs-operator"

# task/env.toml
[chain]
bvsHash = "<your-bvs-hash>"

[owner]
keyringBackend = "os"
keyName = "bvs-user"
```

### Start the System

1. Start Redis and Aggregator

```bash
# Start Redis
sudo systemctl start redis.service

# Run Aggregator
cd aggregator
go run main.go
```

2. Run BVS Off-chain Process (for both operators)

```bash
# Start first operator
cd bvs_offchain
go run main.go

# Start second operator (in new terminal)
cd bvs_offchain
# Update env.toml: change keyName to "bvs-operator-2"
go run main.go
```

3. Start Task Caller

```bash
cd task
go run main.go caller
```

You should see tasks and results appearing after a few seconds. Once you see activity, congratulations! Your BVS is up and running! 🎉

## Conclusion

satRPC demonstrates the potential of Babylon and SatLayer's restaking infrastructure as powerful building blocks for creating decentralized services. By leveraging Bitcoin's security through restaking, we're not just creating a decentralized RPC network - we're showcasing how Babylon Validated Services (BVS) can support diverse distributed validation systems.

The same architecture we use for satRPC can be applied to build various systems requiring distributed validation semantics.

---

Feel free to reach out if you have any questions or want to contribute! 🌐💬
