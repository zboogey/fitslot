package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtIssuer   = "fitslot-api"
	jwtAudience = "fitslot-users"

	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrEmptyJWTSecret   = errors.New("jwt secret cannot be empty")
)

type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func generateToken(userID int, email, role, tokenType, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", ErrEmptyJWTSecret
	}

	now := time.Now()
	expirationTime := now.Add(ttl)

	claims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Audience:  []string{jwtAudience},
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateAccessToken(userID int, email, role, secret string) (string, error) {
	return generateToken(userID, email, role, "access", secret, AccessTokenTTL)
}

func GenerateRefreshToken(userID int, email, role, secret string) (string, error) {
	return generateToken(userID, email, role, "refresh", secret, RefreshTokenTTL)
}

func GenerateTokens(userID int, email, role, accessSecret, refreshSecret string) (accessToken, refreshToken string, err error) {
	accessToken, err = GenerateAccessToken(userID, email, role, accessSecret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = GenerateRefreshToken(userID, email, role, refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func GenerateToken(userID int, email, role, secret string) (string, error) {
	return GenerateAccessToken(userID, email, role, secret)
}

func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	if secret == "" {
		return nil, ErrEmptyJWTSecret
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		},
		jwt.WithIssuer(jwtIssuer),
		jwt.WithAudience(jwtAudience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func RefreshAccessToken(refreshToken, refreshSecret, accessSecret string) (string, *JWTClaims, error) {
	claims, err := ValidateToken(refreshToken, refreshSecret)
	if err != nil {
		return "", nil, err
	}

	if claims.TokenType != "refresh" {
		return "", nil, ErrInvalidTokenType
	}

	newAccessToken, err := GenerateAccessToken(claims.UserID, claims.Email, claims.Role, accessSecret)
	if err != nil {
		return "", nil, err
	}

	return newAccessToken, claims, nil
}
