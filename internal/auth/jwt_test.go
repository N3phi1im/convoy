package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	secret := "test-secret"
	expiry := 24 * time.Hour

	tm := NewTokenManager(secret, expiry)

	require.NotNil(t, tm)
	assert.Equal(t, []byte(secret), tm.secret)
	assert.Equal(t, expiry, tm.expiry)
}

func TestTokenManager_GenerateToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24*time.Hour)
	userID := uuid.New()

	token, err := tm.GenerateToken(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, len(token) > 50, "JWT token should be reasonably long")
}

func TestTokenManager_GenerateToken_NilUserID(t *testing.T) {
	tm := NewTokenManager("test-secret", 24*time.Hour)

	token, err := tm.GenerateToken(uuid.Nil)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestTokenManager_ValidateToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24*time.Hour)
	userID := uuid.New()

	token, err := tm.GenerateToken(userID)
	require.NoError(t, err)

	claims, err := tm.ValidateToken(token)

	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.False(t, claims.ExpiresAt.Time.IsZero())
	assert.False(t, claims.IssuedAt.Time.IsZero())
}

func TestTokenManager_ValidateToken_EmptyToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24*time.Hour)

	claims, err := tm.ValidateToken("")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestTokenManager_ValidateToken_InvalidToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24*time.Hour)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "malformed token",
			token: "not.a.valid.jwt",
		},
		{
			name:  "random string",
			token: "randomstring",
		},
		{
			name:  "empty segments",
			token: "..",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tm.ValidateToken(tt.token)

			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestTokenManager_ValidateToken_WrongSecret(t *testing.T) {
	tm1 := NewTokenManager("secret1", 24*time.Hour)
	tm2 := NewTokenManager("secret2", 24*time.Hour)

	userID := uuid.New()
	token, err := tm1.GenerateToken(userID)
	require.NoError(t, err)

	claims, err := tm2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenManager_ValidateToken_ExpiredToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 1*time.Millisecond)
	userID := uuid.New()

	token, err := tm.GenerateToken(userID)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	claims, err := tm.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestTokenManager_ValidateToken_TamperedToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24*time.Hour)
	userID := uuid.New()

	token, err := tm.GenerateToken(userID)
	require.NoError(t, err)

	tamperedToken := token[:len(token)-5] + "XXXXX"

	claims, err := tm.ValidateToken(tamperedToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestClaims_Structure(t *testing.T) {
	tm := NewTokenManager("test-secret", 24*time.Hour)
	userID := uuid.New()

	token, err := tm.GenerateToken(userID)
	require.NoError(t, err)

	claims, err := tm.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, userID, claims.UserID)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Second)))
	assert.True(t, claims.NotBefore.Time.Before(time.Now().Add(time.Second)))
}

func TestTokenManager_GenerateAndValidate_Integration(t *testing.T) {
	secret := "my-super-secret-key"
	expiry := 1 * time.Hour

	tm := NewTokenManager(secret, expiry)

	testUsers := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.New(),
	}

	for i, userID := range testUsers {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			token, err := tm.GenerateToken(userID)
			require.NoError(t, err)

			claims, err := tm.ValidateToken(token)
			require.NoError(t, err)

			assert.Equal(t, userID, claims.UserID)

			for _, otherUserID := range testUsers {
				if otherUserID != userID {
					assert.NotEqual(t, otherUserID, claims.UserID)
				}
			}
		})
	}
}

func TestTokenManager_TokenExpiry(t *testing.T) {
	tests := []struct {
		name   string
		expiry time.Duration
	}{
		{
			name:   "1 hour",
			expiry: 1 * time.Hour,
		},
		{
			name:   "24 hours",
			expiry: 24 * time.Hour,
		},
		{
			name:   "7 days",
			expiry: 7 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTokenManager("test-secret", tt.expiry)
			userID := uuid.New()

			token, err := tm.GenerateToken(userID)
			require.NoError(t, err)

			claims, err := tm.ValidateToken(token)
			require.NoError(t, err)

			expectedExpiry := time.Now().Add(tt.expiry)
			actualExpiry := claims.ExpiresAt.Time

			timeDiff := actualExpiry.Sub(expectedExpiry)
			assert.True(t, timeDiff < 1*time.Second && timeDiff > -1*time.Second,
				"expiry time should be within 1 second of expected")
		})
	}
}

func TestTokenManager_ValidateToken_WrongSigningMethod(t *testing.T) {
	userID := uuid.New()

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	tm := NewTokenManager("test-secret", 24*time.Hour)
	validatedClaims, err := tm.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}
