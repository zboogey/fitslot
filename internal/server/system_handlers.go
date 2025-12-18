package server

import (
	"net/http"

	"fitslot/internal/api"
	"fitslot/internal/email"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// @Summary      Health check
// @Tags         system
// @Produce      json
// @Success      200 {object} api.HealthResponse
// @Router       /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, api.HealthResponse{Status: "ok"})
}

// @Summary      Queue a test email
// @Tags         system
// @Produce      json
// @Param        email query string true "Recipient email"
// @Success      200 {object} api.MessageResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /test-email [get]
func TestEmail(emailService *email.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		testEmail := c.Query("email")
		if testEmail == "" {
			c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "email parameter required"})
			return
		}

		if err := emailService.Send(c.Request.Context(), testEmail, "Test User", "Test Email from FitSlot", "Email is working!"); err != nil {
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, api.MessageResponse{Message: "Email queued successfully"})
	}
}

// @Summary      Prometheus metrics
// @Description  Exposes Prometheus metrics in text format
// @Tags         system
// @Produce      text/plain
// @Success      200 {string} string
// @Router       /metrics [get]
func Metrics() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
