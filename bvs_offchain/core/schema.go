package core

type Config struct {
	Chain      Chain
	Owner      Owner
	Aggregator Aggregator
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
