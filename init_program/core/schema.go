package core

type Config struct {
	Chain             Chain
	Account           Account
	BvsContract       BvsContract
	StakerOperatorMap []StakerOperatorMap
}

type Chain struct {
	Id  string `json:"id"`
	Rpc string `json:"rpc"`
}

type Account struct {
	KeyDir                 string   `json:"keyDir"`
	KeyringBackend         string   `json:"keyringBackend"`
	Bech32Prefix           string   `json:"bech32Prefix"`
	CallerKeyName          string   `json:"callerKeyName"`
	OperatorsKeyName       []string `json:"operatorsKeyName"`
	UploaderKeyName        string   `json:"uploaderKeyName"`
	AggregatorKeyName      string   `json:"aggregatorKeyName"`
	StakersKeyName         []string `json:"stakersKeyName"`
	ApproverKeyName        string   `json:"approverKeyName"`
	StrategyManagerKeyName string   `json:"strategyManagerKeyName"`
}

type BvsContract struct {
	BvsContractAddr       string `json:"bvsContractAddr"`
	BvsDriverAddr         string `json:"bvsDriverAddr"`
	BvsStateBankAddr      string `json:"stateBankAddr"`
	BvsDirectoryAddr      string `json:"directoryAddr"`
	RewardCoordinatorAddr string `json:"rewardCoordinatorAddr"`
	DelegatorAddr         string `json:"delegationManagerAddr"`
	StrategyMangerAddr    string `json:"strategyMangerAddr"`
	StrategyAddr          string `json:"strategyAddr"`
	Cw20TokenAddr         string `json:"cw20TokenAddr"`
}

type StakerOperatorMap struct {
	StakerKeyName   string `json:"stakerKeyName"`
	Amount          uint64 `json:"amount"`
	OperatorKeyName string `json:"operatorKeyNames"`
}
