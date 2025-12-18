package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Заглушка Claims для теста
type Claims struct {
	UserID    int
	Email     string
	Role      string
	TokenType string
}

func TestAuthMiddlewareHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{"Empty header", "", http.StatusUnauthorized},
		{"Invalid format", "Token abc", http.StatusUnauthorized},
		{"Empty token", "Bearer ", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			c.Request = req

			handler := AuthMiddleware("secret")
			handler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       any
		requiredRole   string
		expectedStatus int
	}{
		{"Correct role", "admin", "admin", http.StatusOK},
		{"Missing role", nil, "admin", http.StatusUnauthorized},
		{"Wrong role type", 123, "admin", http.StatusUnauthorized},
		{"Insufficient role", "user", "admin", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.userRole != nil {
				c.Set("user_role", tt.userRole)
			}
			c.Request = httptest.NewRequest("GET", "/", nil)

			handler := RequireRole(tt.requiredRole)
			handler(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		userID   any
		expected int
		ok       bool
	}{
		{"Valid ID", 42, 42, true},
		{"Missing ID", nil, 0, false},
		{"Wrong type", "abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.userID != nil {
				c.Set("user_id", tt.userID)
			}
			c.Request = httptest.NewRequest("GET", "/", nil)

			id, ok := GetUserID(c)
			assert.Equal(t, tt.expected, id)
			assert.Equal(t, tt.ok, ok)
		})
	}
}
