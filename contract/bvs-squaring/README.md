# BVS Squaring contract

This directory contains the source code for the BVS Squaring smart contract module.

## Development

This contract is developed using Rust and CosmWasm, a powerful framework for developing smart contracts on the Cosmos SDK.

### Prerequisites

- Rust: Ensure you have Rust installed. You can install it from [rustup.rs](https://rustup.rs).
- CosmWasm: Follow the [CosmWasm documentation](https://docs.cosmwasm.com) to set up the environment.

### Building and Testing

To build this contract:

```sh
cargo wasm
```

To test this contract:

```sh
cargo test
```

## Deploying steps for Babylon testnet

A. Build the contract

```sh
cargo wasm
```

B. Install Babylon client:  
Install Golang, access https://go.dev/doc/install to download and install Golang tool for specific OS, then check go installatoin:

```sh
go version
```

Download Babylon client source code and then build and install it:

```sh
git clone https://github.com/babylonchain/babylon

cd babylon

make install
```

Copy babylond executable in babylon build directoy to /user/bin or add babylon/build to PATH environment variable so that babylond can be executed from other pathes on command line:

```sh
cp build/babylond /user/bin/
```

C. Create a new babylon account for deploying contract (if the account does not exist), and transfer some native token of the chain deploying contract to the account.

```sh
babylond keys add wallet
```

D. Upload the built contract wasm code to the chain:

```sh
babylond tx wasm store target/wasm32-unknown-unknown/release/bvs_squaring.wasm --from=wallet --gas=auto --gas-prices=1ubbn --gas-adjustment=1.3 --chain-id=sat-bbn-testnet1 -b=sync --yes --log_format=json --node https://rpc.sat-bbn-testnet1.satlayer.net/
```

E. Get the stored contract code_id:  
Access https://explorer.satlayer.net/satlayer-babylon-testnet/ , search the store transaction with the returned tranaction hash value in previous uploading process, get the code_id attribute in the tx_response json string content.

F. Instantiate the contract with its code_id and init parameters(aggregator is the aggregator account address, state_bank is the state_bank contract address, and bvs_driver is the bvs_driver contract address):

```sh
CODE_ID=28

INIT='{"aggregator": "bbn1yh5vdtu8n55f2e4fjea8gh0dw9gkzv7uxt8jrv", "state_bank": "bbn1h9zjs2zr2xvnpngm9ck8ja7lz2qdt5mcw55ud7wkteycvn7aa4pqpghx2q", "bvs_driver": "bbn18x5lx5dda7896u074329fjk4sflpr65s036gva65m4phavsvs3rqk5e59c"}'


babylond tx wasm instantiate $CODE_ID $INIT --from=wallet --no-admin --label="bvs squaring" --gas=auto --gas-prices=1ubbn --gas-adjustment=1.3 --chain-id=sat-bbn-testnet1 -b=sync --yes --log_format=json --node https://rpc.sat-bbn-testnet1.satlayer.net
```

G. After instantiation, access https://explorer.satlayer.net/satlayer-babylon-testnet/ , search the instantiation transaction with the returned tranaction hash in previous step, get the \_contract_address attribute in the tx_response json string content, it is the instanciated contract address.

### Deployed Contract Addresses on Babylon testnet

- bvs_squaring: bbn1kv4v4aqv6w884myp7x3nkqy5sjf46uacrd6l8zf2yq6rj8mydpssdun4v5

## Deploying steps for Osmosis testnet

A. Build the contract

```sh
cargo wasm
```

B. Install Osmosis client:  
Install Golang, access https://go.dev/doc/install to download and install Golang tool for specific OS, then check go installatoin:

```sh
go version
```

Download Osmosis client source code and then build and install it:

```sh
git clone https://github.com/osmosis-labs/osmosis

cd osmosis

make install
```

Copy osmosis executable in osmosis/build directoy to /user/bin or add osmosis/build to PATH environment variable so that osmosisd can be executed from other pathes on command line:

```sh
cp build/osmosisd /user/bin/
```

C. Create a new ommosis account for deploying contract (if the account does not exist), and transfer some native token of the chain deploying contract to the account.

```sh
osmosis keys add wallet
```

D. Upload the built contract wasm code to the chain:

```sh
osmosisd tx wasm store target/wasm32-unknown-unknown/release/bvs_squaring.wasm --from wallet --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 -y --output json -b async --node https://rpc.testnet.osmosis.zone:443 --chain-id osmo-test-5

```

E. Get the stored contract code_id:  
Access https://celatone.osmosis.zone/, search the store transaction with the returned tranaction hash value in previous uploading process, get the code_id attribute in the Event Logs section.

F. Instantiate the contract with its code_id and init parameters(aggregator is the aggregator account address, state_bank is the state_bank contract address, and bvs_driver is the bvs_driver contract address):

```sh
CODE_ID=11287

INIT='{"aggregator": "osmo1t8jqs8vjltv2lacspvvvw3ygu724jn9s3k4w0r", "state_bank": "osmo14me62ahp32xrkrqnllmsfthfzqxgf0xqshxtk5ghdfwjltdjh2pqdhn8j9", "bvs_driver": "osmo14rrkya0p6h0xf8v3f33grp6dv7lqs2r5xg09zpzjgnggjgfc08fs9kz9ru"}'


osmosisd tx wasm instantiate $CODE_ID "$INIT" --from wallet --label "bvs squaring" --gas-prices 0.025uosmo --gas auto --gas-adjustment 1.3 -b async -y --no-admin --node https://rpc.testnet.osmosis.zone:443 --chain-id osmo-test-5

```

G. After instantiation, access https://celatone.osmosis.zone/ , search the instantiation transaction with the returned tranaction hash in previous step, get the \_contract_address attribute in the Event Logs section, it is the instanciated contract address.

### Deployed Contract Addresses on Osmosis testnet

- bvs_squaring: osmo1spyfwrnzrxfsefek8rus9rt6jhae8c5hx3tghmvf4dzj52ykk8esfpmsgh
