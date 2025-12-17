package server

import (
	"strconv"
	"time"

	"fitslot/internal/metrics"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		
		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			status,
			duration,
		)
	}
}