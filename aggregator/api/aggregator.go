package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/satlayer/hello-world-bvs/aggregator/core"
	"github.com/satlayer/hello-world-bvs/aggregator/svc"
	"github.com/satlayer/hello-world-bvs/aggregator/util"
	"github.com/satlayer/satlayer-api/signer"
)

type Payload struct {
	TaskId    uint64 `json:"taskID" binding:"required"`
	Result    string `json:"result" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	PubKey    string `json:"pubKey" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

// Aggregator handles the aggregator endpoint for the API.
//
// It parses the payload from the request body and verifies the signature.
// It checks if the timestamp is within the allowed range.
// It verifies if the task is finished and if the operator has already sent the task.
// If all checks pass, it saves the task to the queue.
// It returns an HTTP response with the status of the operation.
//
// Parameters:
// - c: The gin.Context object representing the HTTP request and response.
//
// Returns:
// - None.
func Aggregator(c *gin.Context) {
	var payload Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Role != core.RolePerformer && payload.Role != core.RoleAttester {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	nowTs := time.Now().Unix()
	if payload.Timestamp > nowTs || payload.Timestamp < nowTs-60*2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp out of range"})
		return
	}

	pubKey, address, err := util.PubKeyToAddress(payload.PubKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate result format based on role
	if payload.Role == core.RolePerformer {
		resultParts := strings.Split(payload.Result, "-")
		if len(resultParts) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result format for performer"})
			return
		}
		latestBlockNumber := resultParts[0]
		latestBlockHash := resultParts[1]
		fmt.Printf("Latest Block Number: %s\n", latestBlockNumber)
		fmt.Printf("Latest Block Hash: %s\n", latestBlockHash)
	} else {
		if payload.Result != "true" && payload.Result != "false" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result format for attester"})
			return
		}
		fmt.Printf("Attester validation result: %s\n", payload.Result)
	}

	msgPayload := fmt.Sprintf("%s-%d-%d-%s", core.C.Chain.BvsHash, payload.Timestamp, payload.TaskId, payload.Result)
	msgBytes := []byte(msgPayload)
	if isValid, err := signer.VerifySignature(pubKey, msgBytes, payload.Signature); err != nil || !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	pkTaskFinished := fmt.Sprintf("%s%d", core.PkTaskFinished, payload.TaskId)
	if isExist, err := core.S.RedisConn.Exists(c, pkTaskFinished).Result(); err != nil || isExist == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task already finished"})
		return
	}

	if ok, err := svc.MONITOR.VerifyOperator(address); err != nil || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid operator"})
		return
	}

	verificationKey := fmt.Sprintf("%s%d", core.PkTaskVerification, payload.TaskId)
	var taskVerification core.TaskVerification

	existingData, err := core.S.RedisConn.Get(c, verificationKey).Result()
	if err != nil && err != redis.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get verification data"})
		return
	}

	if existingData != "" {
		if err := json.Unmarshal([]byte(existingData), &taskVerification); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse verification data"})
			return
		}
	} else {
		taskVerification = core.TaskVerification{
			Attesters: make(map[string]*core.TaskSubmission),
		}
	}

	submission := &core.TaskSubmission{
		Address:   address,
		Result:    payload.Result,
		Timestamp: payload.Timestamp,
		Role:      payload.Role,
	}

	if payload.Role == core.RolePerformer {
		if taskVerification.Performer != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "performer already submitted"})
			return
		}
		taskVerification.Performer = submission
	} else {
		if _, exists := taskVerification.Attesters[address]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "attester already submitted"})
			return
		}
		taskVerification.Attesters[address] = submission
	}

	updatedData, err := json.Marshal(taskVerification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal verification data"})
		return
	}

	if err := core.S.RedisConn.Set(c, verificationKey, updatedData, 24*time.Hour).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save verification data"})
		return
	}

	// Only process the task if we have both performer and enough attesters
	if taskVerification.Performer != nil && len(taskVerification.Attesters) >= 1 {
		// Count positive validations
		positiveVotes := 0
		// _totalVotes := len(taskVerification.Attesters)

		for _, attester := range taskVerification.Attesters {
			if attester.Result == "true" {
				positiveVotes++
			}
		}

		// For now, always set result to 1 (mocked verification)
		task := core.Task{
			TaskId: payload.TaskId,
			TaskResult: core.TaskResult{
				Operator: taskVerification.Performer.Address,
				Result:   1, // Hardcoded to 1 for now
			},
		}

		taskStr, err := json.Marshal(task)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if _, err := core.S.RedisConn.LPush(c, core.PkTaskQueue, taskStr).Result(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := core.S.RedisConn.Set(c, pkTaskFinished, "1", 24*time.Hour).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark task as finished"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
