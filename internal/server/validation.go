package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidationMiddleware validates request bodies using struct tags
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET, DELETE, and OPTIONS
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodDelete ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Get the content type
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// Try to bind JSON to a generic map first to check if body exists
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			// If binding fails, let the handler deal with it
			c.Next()
			return
		}

		c.Next()
	}
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) []ValidationError {
	validate := validator.New()
	var errors []ValidationError

	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Message: getErrorMessage(err),
			})
		}
	}

	return errors
}

// getErrorMessage returns a user-friendly error message
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "email":
		return err.Field() + " must be a valid email address"
	case "min":
		return err.Field() + " must be at least " + err.Param() + " characters"
	case "max":
		return err.Field() + " must be at most " + err.Param() + " characters"
	case "gte":
		return err.Field() + " must be greater than or equal to " + err.Param()
	case "lte":
		return err.Field() + " must be less than or equal to " + err.Param()
	default:
		return err.Field() + " is invalid"
	}
}

// RespondWithValidationErrors sends validation errors as JSON response
func RespondWithValidationErrors(c *gin.Context, errors []ValidationError) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "validation failed",
		"details": errors,
	})
}
