package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/satlayer/hello-world-bvs/aggregator/core"
)

func GetTaskData(c *gin.Context) {
	taskId := c.Param("taskId")

	verificationKey := fmt.Sprintf("%s%s", core.PkTaskVerification, taskId)
	var taskVerification core.TaskVerification

	existingData, err := core.S.RedisConn.Get(c, verificationKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task data not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get verification data"})
		return
	}

	if err := json.Unmarshal([]byte(existingData), &taskVerification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse verification data"})
		return
	}

	// If no performer data yet, return not found
	if taskVerification.Performer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "performer data not found"})
		return
	}

	// Return performer's data
	c.JSON(http.StatusOK, gin.H{
		"result":  taskVerification.Performer.Result,
		"address": taskVerification.Performer.Address,
	})
}
