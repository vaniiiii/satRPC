package uploader

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/satlayer/satlayer-api/chainio/api"
	"github.com/satlayer/satlayer-api/chainio/types"
)

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  string
}

type HashResponse struct {
	HashBinary []byte `json:"hash_binary"`
}

type MerkleizeLeavesResponse struct {
	RootHashBinary []byte `json:"root_hash_binary"`
}

func (u *Uploader) rpcSubmission(rewards []Submission) error {
	const calcInterval = 86_400 // 1 day
	now := time.Now().Unix()
	startTime := now - now%calcInterval
	submissions := make([]types.RewardsSubmission, 0)
	for _, reward := range rewards {
		submissions = append(submissions, types.RewardsSubmission{
			StrategiesAndMultipliers: []types.StrategyAndMultiplier{{
				Strategy:   reward.Strategy,
				Multiplier: 1,
			}},
			Token:          reward.Token,
			Amount:         strconv.FormatFloat(math.Floor(reward.Amount), 'f', -1, 64),
			StartTimestamp: fmt.Sprintf("%d000000000", startTime),
			Duration:       calcInterval,
		})
	}
	fmt.Printf("submissions: %+v\n", submissions)
	resp, err := u.rewardsCoordinator.CreateRewardsForAllSubmission(context.Background(), submissions)
	fmt.Println("CreateRewardsForAllSubmission txn hash: ", resp.Hash.String())
	return err
}

func (u *Uploader) rpcTokenHash(token *TokenAmount) (string, error) {
	resp, err := u.rewardsCoordinator.CalculateTokenLeafHash(token.Token, token.RewardAmount)
	if err != nil {
		fmt.Println("CalculateTokenLeafHash err: ", err)
		return "", err
	}
	var hashResponse HashResponse
	if err := json.Unmarshal(resp.Data, &hashResponse); err != nil {
		fmt.Println("unmarshal err: ", err)
		return "", err
	}
	hashStr := base64.StdEncoding.EncodeToString(hashResponse.HashBinary)
	return hashStr, err
}

func (u *Uploader) rpcMerkleizeLeaves(leaves []string) (string, error) {
	resp, err := u.rewardsCoordinator.MerkleizeLeaves(leaves)
	if err != nil {
		fmt.Println("merkleizeLeaves err: ", err)
		return "", err
	}
	var merkleizeLeavesResponse MerkleizeLeavesResponse
	if err := json.Unmarshal(resp.Data, &merkleizeLeavesResponse); err != nil {
		fmt.Println("unmarshal err: ", err)
		return "", err
	}
	merkleRoot := base64.StdEncoding.EncodeToString(merkleizeLeavesResponse.RootHashBinary)
	return merkleRoot, err
}

type EarnerLeafHashResponse struct {
	RootHashBinary []byte `json:"hash_binary"`
}

type UnderlyingTokenResponse struct {
	UnderlyingTokenAddr string `json:"underlying_token_addr"`
}

func (u *Uploader) rpcEarnerLeafHash(staker, rootHash string) (string, error) {
	resp, err := u.rewardsCoordinator.CalculateEarnerLeafHash(staker, rootHash)
	if err != nil {
		fmt.Println("CalculateEarnerLeafHash err: ", err)
		return "", err
	}
	var earnerLeafHashResponse EarnerLeafHashResponse
	err = json.Unmarshal(resp.Data, &earnerLeafHashResponse)
	hashStr := base64.StdEncoding.EncodeToString(earnerLeafHashResponse.RootHashBinary)
	return hashStr, err
}

func (u *Uploader) rpcUnderlyingToken(strategy string) (string, error) {
	strategyBase := api.NewStrategyBase(u.chainIO)
	strategyBase.BindClient(strategy)
	resp, err := strategyBase.UnderlyingToken()
	if err != nil {
		return "", err
	}
	var tokenRsp UnderlyingTokenResponse
	err = json.Unmarshal(resp.Data, &tokenRsp)

	return tokenRsp.UnderlyingTokenAddr, nil
}

func (u *Uploader) rpcSubmitHashRoot(rootHash string) error {
	timestamp := time.Now().Unix() - 3600
	rsp, err := u.rewardsCoordinator.SubmitRoot(context.Background(), rootHash, uint64(timestamp))
	if err != nil {
		fmt.Println("SubmitRootHash err: ", err)
		return err
	}
	fmt.Println("SubmitRootHash txn hash: ", rsp.Hash.String())
	return err
}
