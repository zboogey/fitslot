package user

import (
	"net/http"

	"fitslot/internal/api"
	"fitslot/internal/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   Service
	jwtSecret string
}

func NewHandler(service Service, jwtSecret string) *Handler {
	return &Handler{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

// @Summary      Register user
// @Description  Register a new user (member role by default) and return access/refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body user.RegisterRequest true "Register payload"
// @Success      201 {object} user.LoginResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      409 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	user, accessToken, refreshToken, err := h.service.Register(ctx, req)
	if err != nil {
		switch err {
		case ErrEmailExists:
			c.JSON(http.StatusConflict, api.ErrorResponse{Error: "Email already registered"})
		default:
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	})
}

// @Summary      Login
// @Description  Login with email and password to receive access/refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body user.LoginRequest true "Login payload"
// @Success      200 {object} user.LoginResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	user, accessToken, refreshToken, err := h.service.Login(ctx, req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	})
}

// @Summary      Get current user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} user.User
// @Failure      401 {object} api.ErrorResponse
// @Failure      404 {object} api.ErrorResponse
// @Router       /me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "User not authenticated"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.service.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary      Refresh access token
// @Description  Exchange a refresh token for a new access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body user.RefreshTokenRequest true "Refresh token payload"
// @Success      200 {object} user.RefreshTokenResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Router       /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	newAccessToken, user, err := h.service.RefreshToken(ctx, req.RefreshToken, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken: newAccessToken,
		User:        *user,
	})
}
