package core

import "time"

type Config struct {
	Chain      Chain
	Owner      Owner
	Aggregator Aggregator
	Rpc        Rpc
}

type Chain struct {
	Id           string `json:"id"`
	Rpc          string `json:"rpc"`
	BvsHash      string `json:"bvsHash"`
	BvsDirectory string `json:"bvsDirectory"`
	BvsDriver    string `json:"bvsDriver"`
	StateBank    string `json:"stateBank"`
}

type Owner struct {
	KeyDir         string `json:"keyDir"`
	KeyName        string `json:"keyName"`
	KeyringBackend string `json:"keyringBackend"`
	Bech32Prefix   string `json:"bech32Prefix"`
}

type Aggregator struct {
	Url string `json:"url"`
}

type Rpc struct {
	Endpoint string `json:endpoint`
}

type StatusResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		SyncInfo struct {
			LatestBlockHeight string    `json:"latest_block_height"`
			LatestBlockHash   string    `json:"latest_block_hash"`
			LatestBlockTime   time.Time `json:"latest_block_time"`
		} `json:"sync_info"`
	} `json:"result"`
}

type BlockResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		BlockID struct {
			Hash string `json:"hash"`
		} `json:"block_id"`
		Block struct {
			Header struct {
				Height  string    `json:"height"`
				Time    time.Time `json:"time"`
				ChainID string    `json:"chain_id"`
			} `json:"header"`
		} `json:"block"`
	} `json:"result"`
}
