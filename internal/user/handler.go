package user

import (
    "net/http"

    "fitslot/internal/auth"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
)

type Handler struct {
    repo      *Repository
    jwtSecret string
}

func NewHandler(db *sqlx.DB, jwtSecret string) *Handler {
    return &Handler{
        repo:      NewRepository(db),
        jwtSecret: jwtSecret,
    }
}

// Register godoc
// @Summary      Register new user
// @Description  Creates a new member user and returns access & refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "User registration data"
// @Success      201      {object}  LoginResponse
// @Failure      400      {object}  gin.H
// @Failure      409      {object}  gin.H
// @Failure      500      {object}  gin.H
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    exists, err := h.repo.EmailExists(req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if exists {
        c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
        return
    }

    passwordHash, err := auth.HashPassword(req.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    user, err := h.repo.Create(req.Name, req.Email, passwordHash, "member")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    accessToken, refreshToken, err := auth.GenerateTokens(
        user.ID,
        user.Email,
        user.Role,
        h.jwtSecret,
        h.jwtSecret,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
        return
    }

    c.JSON(http.StatusCreated, LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        User:         *user,
    })
}

// Login godoc
// @Summary      Login user
// @Description  Authenticates user by email and password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "User credentials"
// @Success      200      {object}  LoginResponse
// @Failure      400      {object}  gin.H
// @Failure      401      {object}  gin.H
// @Failure      500      {object}  gin.H
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.repo.FindByEmail(req.Email)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    if !auth.CheckPassword(user.PasswordHash, req.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    accessToken, refreshToken, err := auth.GenerateTokens(
        user.ID,
        user.Email,
        user.Role,
        h.jwtSecret,
        h.jwtSecret,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
        return
    }

    c.JSON(http.StatusOK, LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        User:         *user,
    })
}

// GetMe godoc
// @Summary      Get current user
// @Description  Returns profile of the authenticated user.
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  User
// @Failure      401  {object}  gin.H
// @Failure      404  {object}  gin.H
// @Router       /me [get]
func (h *Handler) GetMe(c *gin.Context) {
    userID, exists := auth.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    user, err := h.repo.FindByID(userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, user)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Returns new access token using a valid refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      map[string]string  true  "Refresh token payload"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  gin.H
// @Failure      401      {object}  gin.H
// @Failure      404      {object}  gin.H
// @Failure      500      {object}  gin.H
// @Router       /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
        return
    }

    _, claims, err := auth.RefreshAccessToken(req.RefreshToken, h.jwtSecret, h.jwtSecret)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
        return
    }

    user, err := h.repo.FindByID(claims.UserID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    newAccessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, h.jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "access_token": newAccessToken,
        "user":         user,
    })
}
