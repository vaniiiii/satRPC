package core

import "github.com/go-redis/redis/v8"

type Config struct {
	Chain    Chain
	Owner    Owner
	Database Database
	Reward   Reward
}

type Chain struct {
	Id                string `json:"id"`
	Rpc               string `json:"rpc"`
	InitBlockNum      uint64 `json:"initBlockNum"`
	RewardCoordinator string `json:"rewardCoordinator"`
	BvsHash           string `json:"bvsHash"`
	BvsDirectory      string `json:"bvsDirectory"`
	DelegationManager string `json:"delegationManager"`
}

type Owner struct {
	KeyDir         string `json:"keyDir"`
	KeyName        string `json:"keyName"`
	KeyringBackend string `json:"keyringBackend"`
	Bech32Prefix   string `json:"bech32Prefix"`
}

type Reward struct {
	Amount           float64 `json:"amount"`
	OperatorRatio    float64 `json:"operatorRatio"`
	OperatorStrategy string  `json:"operatorStrategy"`
}

type Database struct {
	RedisHost     string `json:"redisHost"`
	RedisPassword string `json:"redisPassword"`
	RedisDb       int    `json:"redisDb"`
}

type Store struct {
	RedisConn *redis.Client
}
