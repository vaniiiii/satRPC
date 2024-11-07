package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

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

	resultParts := strings.Split(payload.Result, "-")
	if len(resultParts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result format"})
		return
	}

	// @reminder Add verification here
	latestBlockNumber := resultParts[0]
	latestBlockHash := resultParts[1]
	fmt.Printf("Latest Block Number: %s\n", latestBlockNumber)
	fmt.Printf("Latest Block Hash: %s\n", latestBlockHash)

	// For now let's hardocde it to 1(true)
	var verificationResult int64 = 1

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

	taskOperatorKey := fmt.Sprintf("%s%d", core.PkTaskOperator, payload.TaskId)
	if result, err := core.S.RedisConn.Eval(c, core.LuaScript, []string{taskOperatorKey}, address).Result(); err != nil || result.(int64) == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task already send"})
		return
	}

	task := core.Task{TaskId: payload.TaskId, TaskResult: core.TaskResult{Operator: address, Result: verificationResult}}
	taskStr, err := json.Marshal(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := core.S.RedisConn.LPush(c, core.PkTaskQueue, taskStr).Result(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
