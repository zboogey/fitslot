package server

import (
	"time"

	"fitslot/internal/logger"

	"github.com/gin-gonic/gin"
)

// RequestLoggingMiddleware logs HTTP requests with structured logging
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP request",
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", clientIP,
			"user_agent", c.Request.UserAgent(),
		)
	}
}
