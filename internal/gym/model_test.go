package gym

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateGymRequest_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/", func(c *gin.Context) {
		var req CreateGymRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	w := httptest.NewRecorder()
	reqBody := bytes.NewBuffer([]byte(`{}`))
	req, _ := http.NewRequest(http.MethodPost, "/", reqBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Name")
	assert.Contains(t, w.Body.String(), "required")
}

func TestCreateTimeSlotRequest_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/", func(c *gin.Context) {
		var req CreateTimeSlotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	w := httptest.NewRecorder()
	reqBody := bytes.NewBuffer([]byte(`{}`))
	req, _ := http.NewRequest(http.MethodPost, "/", reqBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "StartTime")
	assert.Contains(t, w.Body.String(), "required")
	assert.Contains(t, w.Body.String(), "EndTime")
	assert.Contains(t, w.Body.String(), "required")
	assert.Contains(t, w.Body.String(), "Capacity")
	assert.Contains(t, w.Body.String(), "required")
}
