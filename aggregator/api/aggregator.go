package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/satlayer/hello-world-bvs/aggregator/core"
	"github.com/satlayer/hello-world-bvs/aggregator/svc"
	"github.com/satlayer/hello-world-bvs/aggregator/util"
	"github.com/satlayer/satlayer-api/signer"
)

type Payload struct {
	TaskId    uint64 `json:"taskID" binding:"required"`
	Result    int64  `json:"result" binding:"required"`
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
	// parse payload
	var payload Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("payload: %+v\n", payload)

	// get current timestamp
	nowTs := time.Now().Unix()
	if payload.Timestamp > nowTs || payload.Timestamp < nowTs-60*2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp out of range"})
		return
	}

	// verify signature
	pubKey, address, err := util.PubKeyToAddress(payload.PubKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("pubKey: %s\n", pubKey)
	fmt.Printf("address: %s\n", address)
	fmt.Printf("payload.PubKey: %s\n", payload.PubKey)

	msgPayload := fmt.Sprintf("%s-%d-%d-%d", core.C.Chain.BvsHash, payload.Timestamp, payload.TaskId, payload.Result)
	msgBytes := []byte(msgPayload)
	if isValid, err := signer.VerifySignature(pubKey, msgBytes, payload.Signature); err != nil || !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	// verify task is finished
	pkTaskFinished := fmt.Sprintf("%s%d", core.PkTaskFinished, payload.TaskId)
	if isExist, err := core.S.RedisConn.Exists(c, pkTaskFinished).Result(); err != nil || isExist == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task already finished"})
		return
	}

	if ok, err := svc.MONITOR.VerifyOperator(address); err != nil || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid operator"})
		return
	}

	// verify operator is already send
	taskOperatorKey := fmt.Sprintf("%s%d", core.PkTaskOperator, payload.TaskId)
	if result, err := core.S.RedisConn.Eval(c, core.LuaScript, []string{taskOperatorKey}, address).Result(); err != nil || result.(int64) == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task already send"})
		return
	}

	// save task to queue
	task := core.Task{TaskId: payload.TaskId, TaskResult: core.TaskResult{Operator: address, Result: payload.Result}}
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
