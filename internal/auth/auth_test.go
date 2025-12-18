package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-12345"

func TestHashPassword(t *testing.T) {
	t.Run("Successfully hash password", func(t *testing.T) {
		password := "mySecurePassword123"
		hashed, err := HashPassword(password)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)
	})

	t.Run("Different hashes for same password", func(t *testing.T) {
		password := "samePassword"
		hash1, _ := HashPassword(password)
		hash2, _ := HashPassword(password)
		
		// Bcrypt генерирует разные хеши для одного пароля (из-за соли)
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestCheckPassword(t *testing.T) {
	password := "correctPassword"
	hashed, _ := HashPassword(password)

	t.Run("Correct password", func(t *testing.T) {
		result := CheckPassword(hashed, password)
		assert.True(t, result)
	})

	t.Run("Incorrect password", func(t *testing.T) {
		result := CheckPassword(hashed, "wrongPassword")
		assert.False(t, result)
	})

	t.Run("Empty password", func(t *testing.T) {
		result := CheckPassword(hashed, "")
		assert.False(t, result)
	})
}

func TestGenerateAccessToken(t *testing.T) {
	t.Run("Successfully generate access token", func(t *testing.T) {
		token, err := GenerateAccessToken(1, "user@example.com", "user", testSecret)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Fail with empty secret", func(t *testing.T) {
		token, err := GenerateAccessToken(1, "user@example.com", "user", "")
		
		assert.Error(t, err)
		assert.Equal(t, ErrEmptyJWTSecret, err)
		assert.Empty(t, token)
	})

	t.Run("Token contains correct claims", func(t *testing.T) {
		userID := 42
		email := "test@example.com"
		role := "admin"

		token, err := GenerateAccessToken(userID, email, role, testSecret)
		require.NoError(t, err)

		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, "access", claims.TokenType)
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	t.Run("Successfully generate refresh token", func(t *testing.T) {
		token, err := GenerateRefreshToken(1, "user@example.com", "user", testSecret)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Refresh token has longer expiration", func(t *testing.T) {
		token, err := GenerateRefreshToken(1, "user@example.com", "user", testSecret)
		require.NoError(t, err)

		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)

		assert.Equal(t, "refresh", claims.TokenType)
		
		// Проверяем что токен истекает примерно через 7 дней
		expectedExpiry := time.Now().Add(RefreshTokenTTL)
		actualExpiry := claims.ExpiresAt.Time
		
		diff := actualExpiry.Sub(expectedExpiry).Abs()
		assert.Less(t, diff, 2*time.Second) // допуск 2 секунды
	})
}

func TestGenerateTokens(t *testing.T) {
	accessSecret := "access-secret"
	refreshSecret := "refresh-secret"

	t.Run("Successfully generate both tokens", func(t *testing.T) {
		accessToken, refreshToken, err := GenerateTokens(
			1,
			"user@example.com",
			"user",
			accessSecret,
			refreshSecret,
		)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotEqual(t, accessToken, refreshToken)
	})

	t.Run("Fail with empty access secret", func(t *testing.T) {
		accessToken, refreshToken, err := GenerateTokens(
			1,
			"user@example.com",
			"user",
			"",
			refreshSecret,
		)
		
		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

	t.Run("Fail with empty refresh secret", func(t *testing.T) {
		accessToken, refreshToken, err := GenerateTokens(
			1,
			"user@example.com",
			"user",
			accessSecret,
			"",
		)
		
		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})
}

func TestValidateToken(t *testing.T) {
	userID := 100
	email := "test@example.com"
	role := "admin"

	t.Run("Successfully validate valid token", func(t *testing.T) {
		token, _ := GenerateAccessToken(userID, email, role, testSecret)
		
		claims, err := ValidateToken(token, testSecret)
		
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("Fail with empty secret", func(t *testing.T) {
		token, _ := GenerateAccessToken(userID, email, role, testSecret)
		
		claims, err := ValidateToken(token, "")
		
		assert.Error(t, err)
		assert.Equal(t, ErrEmptyJWTSecret, err)
		assert.Nil(t, claims)
	})

	t.Run("Fail with wrong secret", func(t *testing.T) {
		token, _ := GenerateAccessToken(userID, email, role, testSecret)
		
		claims, err := ValidateToken(token, "wrong-secret")
		
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Fail with invalid token format", func(t *testing.T) {
		claims, err := ValidateToken("invalid.token.format", testSecret)
		
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Fail with expired token", func(t *testing.T) {
		// Создаем токен с истекшим сроком
		now := time.Now()
		pastTime := now.Add(-1 * time.Hour)
		
		claims := &JWTClaims{
			UserID:    userID,
			Email:     email,
			Role:      role,
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtIssuer,
				Audience:  []string{jwtAudience},
				ExpiresAt: jwt.NewNumericDate(pastTime),
				IssuedAt:  jwt.NewNumericDate(pastTime.Add(-15 * time.Minute)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(testSecret))
		
		validatedClaims, err := ValidateToken(tokenString, testSecret)
		
		assert.Error(t, err)
		assert.Equal(t, ErrTokenExpired, err)
		assert.Nil(t, validatedClaims)
	})

	t.Run("Token has correct issuer and audience", func(t *testing.T) {
		token, _ := GenerateAccessToken(userID, email, role, testSecret)
		
		claims, err := ValidateToken(token, testSecret)
		
		require.NoError(t, err)
		assert.Equal(t, jwtIssuer, claims.Issuer)
		assert.Contains(t, claims.Audience, jwtAudience)
	})
}

func TestRefreshAccessToken(t *testing.T) {
	accessSecret := "access-secret"
	refreshSecret := "refresh-secret"
	userID := 1
	email := "user@example.com"
	role := "user"

	t.Run("Successfully refresh access token", func(t *testing.T) {
		refreshToken, _ := GenerateRefreshToken(userID, email, role, refreshSecret)
		
		newAccessToken, claims, err := RefreshAccessToken(refreshToken, refreshSecret, accessSecret)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("Fail with access token instead of refresh token", func(t *testing.T) {
		accessToken, _ := GenerateAccessToken(userID, email, role, accessSecret)
		
		newAccessToken, claims, err := RefreshAccessToken(accessToken, accessSecret, accessSecret)
		
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTokenType, err)
		assert.Empty(t, newAccessToken)
		assert.Nil(t, claims)
	})

	t.Run("Fail with invalid refresh token", func(t *testing.T) {
		newAccessToken, claims, err := RefreshAccessToken("invalid.token", refreshSecret, accessSecret)
		
		assert.Error(t, err)
		assert.Empty(t, newAccessToken)
		assert.Nil(t, claims)
	})

	t.Run("Fail with expired refresh token", func(t *testing.T) {
		// Создаем истекший refresh токен
		now := time.Now()
		pastTime := now.Add(-8 * 24 * time.Hour) // 8 дней назад
		
		claims := &JWTClaims{
			UserID:    userID,
			Email:     email,
			Role:      role,
			TokenType: "refresh",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtIssuer,
				Audience:  []string{jwtAudience},
				ExpiresAt: jwt.NewNumericDate(pastTime),
				IssuedAt:  jwt.NewNumericDate(pastTime.Add(-7 * 24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		expiredRefreshToken, _ := token.SignedString([]byte(refreshSecret))
		
		newAccessToken, validatedClaims, err := RefreshAccessToken(expiredRefreshToken, refreshSecret, accessSecret)
		
		assert.Error(t, err)
		assert.Equal(t, ErrTokenExpired, err)
		assert.Empty(t, newAccessToken)
		assert.Nil(t, validatedClaims)
	})

	t.Run("New access token is valid", func(t *testing.T) {
		refreshToken, _ := GenerateRefreshToken(userID, email, role, refreshSecret)
		
		newAccessToken, _, err := RefreshAccessToken(refreshToken, refreshSecret, accessSecret)
		require.NoError(t, err)
		
		// Валидируем новый access token
		accessClaims, err := ValidateToken(newAccessToken, accessSecret)
		
		assert.NoError(t, err)
		assert.Equal(t, userID, accessClaims.UserID)
		assert.Equal(t, "access", accessClaims.TokenType)
	})
}

func TestGenerateToken(t *testing.T) {
	t.Run("GenerateToken is alias for GenerateAccessToken", func(t *testing.T) {
		token1, err1 := GenerateToken(1, "user@example.com", "user", testSecret)
		token2, err2 := GenerateAccessToken(1, "user@example.com", "user", testSecret)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEmpty(t, token1)
		assert.NotEmpty(t, token2)
		
		// Оба токена должны быть валидными access токенами
		claims1, _ := ValidateToken(token1, testSecret)
		claims2, _ := ValidateToken(token2, testSecret)
		
		assert.Equal(t, "access", claims1.TokenType)
		assert.Equal(t, "access", claims2.TokenType)
	})
}

func TestTokenExpiration(t *testing.T) {
	t.Run("Access token expires after 15 minutes", func(t *testing.T) {
		token, err := GenerateAccessToken(1, "user@example.com", "user", testSecret)
		require.NoError(t, err)

		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)

		expectedExpiry := time.Now().Add(AccessTokenTTL)
		actualExpiry := claims.ExpiresAt.Time
		
		diff := actualExpiry.Sub(expectedExpiry).Abs()
		assert.Less(t, diff, 2*time.Second)
	})

	t.Run("Refresh token expires after 7 days", func(t *testing.T) {
		token, err := GenerateRefreshToken(1, "user@example.com", "user", testSecret)
		require.NoError(t, err)

		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)

		expectedExpiry := time.Now().Add(RefreshTokenTTL)
		actualExpiry := claims.ExpiresAt.Time
		
		diff := actualExpiry.Sub(expectedExpiry).Abs()
		assert.Less(t, diff, 2*time.Second)
	})
}

func TestJWTClaimsStructure(t *testing.T) {
	t.Run("Claims contain all required fields", func(t *testing.T) {
		userID := 123
		email := "test@example.com"
		role := "moderator"
		
		token, err := GenerateAccessToken(userID, email, role, testSecret)
		require.NoError(t, err)
		
		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)
		
		// Проверяем все поля
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, "access", claims.TokenType)
		assert.Equal(t, jwtIssuer, claims.Issuer)
		assert.Contains(t, claims.Audience, jwtAudience)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
	})
}