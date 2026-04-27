package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "securepassword123",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "abc",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "this_is_a_very_long_password_with_many_characters_1234567890",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash)
				assert.True(t, len(hash) > 50, "hash should be reasonably long")
			}
		})
	}
}

func TestHashPassword_Uniqueness(t *testing.T) {
	password := "testpassword"

	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes (salt)")
}

func TestCheckPassword(t *testing.T) {
	password := "correctpassword"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
		{
			name:     "invalid hash",
			password: password,
			hash:     "not_a_valid_bcrypt_hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.password, tt.hash)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestPasswordHashAndCheck_Integration(t *testing.T) {
	testCases := []string{
		"simple",
		"with spaces",
		"with-special-chars!@#$%^&*()",
		"unicode-密码-🔒",
		"LongPasswordButWithin72ByteLimit1234567890",
	}

	for _, password := range testCases {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			require.NoError(t, err)

			assert.True(t, CheckPassword(password, hash), "correct password should validate")
			assert.False(t, CheckPassword(password+"wrong", hash), "incorrect password should not validate")
		})
	}
}
