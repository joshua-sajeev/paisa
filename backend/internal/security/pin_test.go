package security_test

import (
	"errors"
	"testing"

	"github.com/joshu-sajeev/paisa/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePIN(t *testing.T) {
	tests := []struct {
		name    string
		pin     string
		wantErr bool
	}{
		// Valid.
		{name: "valid 123457", pin: "123457"},
		{name: "valid 246802", pin: "246802"},
		{name: "valid 135792", pin: "135792"},
		{name: "valid 192837", pin: "192837"},
		{name: "valid 654329", pin: "654329"},
		{name: "valid 481516", pin: "481516"},
		{name: "valid 739201", pin: "739201"},
		{name: "valid 506172", pin: "506172"},
		{name: "valid 918273", pin: "918273"},

		// Invalid format.
		{name: "too short", pin: "12345", wantErr: true},
		{name: "too long", pin: "1234567", wantErr: true},
		{name: "contains letter", pin: "12345a", wantErr: true},
		{name: "contains space", pin: "123 456", wantErr: true},
		{name: "contains symbol", pin: "-12345", wantErr: true},
		{name: "contains decimal", pin: "123.456", wantErr: true},

		// Weak.
		{name: "all zeros", pin: "000000", wantErr: true},
		{name: "all ones", pin: "111111", wantErr: true},
		{name: "all twos", pin: "222222", wantErr: true},
		{name: "all threes", pin: "333333", wantErr: true},
		{name: "all fours", pin: "444444", wantErr: true},
		{name: "all fives", pin: "555555", wantErr: true},
		{name: "all sixes", pin: "666666", wantErr: true},
		{name: "all sevens", pin: "777777", wantErr: true},
		{name: "all eights", pin: "888888", wantErr: true},
		{name: "all nines", pin: "999999", wantErr: true},

		// Increasing sequential.
		{name: "sequential 012345", pin: "012345", wantErr: true},
		{name: "sequential 123456", pin: "123456", wantErr: true},
		{name: "sequential 234567", pin: "234567", wantErr: true},
		{name: "sequential 345678", pin: "345678", wantErr: true},
		{name: "sequential 456789", pin: "456789", wantErr: true},
		{name: "sequential 567890", pin: "567890", wantErr: true},
		{name: "sequential 678901", pin: "678901", wantErr: true},
		{name: "sequential 789012", pin: "789012", wantErr: true},
		{name: "sequential 890123", pin: "890123", wantErr: true},
		{name: "sequential 901234", pin: "901234", wantErr: true},

		// Decreasing sequential.
		{name: "sequential 987654", pin: "987654", wantErr: true},
		{name: "sequential 876543", pin: "876543", wantErr: true},
		{name: "sequential 765432", pin: "765432", wantErr: true},
		{name: "sequential 654321", pin: "654321", wantErr: true},
		{name: "sequential 543210", pin: "543210", wantErr: true},
		{name: "sequential 432109", pin: "432109", wantErr: true},
		{name: "sequential 321098", pin: "321098", wantErr: true},
		{name: "sequential 210987", pin: "210987", wantErr: true},
		{name: "sequential 109876", pin: "109876", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidatePIN(tt.pin)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, security.ErrInvalidPIN))
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestHashPIN(t *testing.T) {
	tests := []struct {
		name    string
		pin     string
		wantErr bool
	}{
		{
			name: "valid PIN",
			pin:  "481516",
		},
		{
			name:    "invalid PIN",
			pin:     "123456",
			wantErr: true,
		},
		{
			name:    "short PIN",
			pin:     "12345",
			wantErr: true,
		},
		{
			name:    "non-numeric PIN",
			pin:     "12345a",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := security.HashPIN(tt.pin)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, hash)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.Contains(t, hash, "$argon2id$")
			assert.Contains(t, hash, "v=19")
		})
	}
}

func TestHashPIN_DifferentHashForSamePIN(t *testing.T) {
	pin := "481516"

	hash1, err := security.HashPIN(pin)
	require.NoError(t, err)

	hash2, err := security.HashPIN(pin)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestVerifyPIN(t *testing.T) {
	pin := "481516"

	hash, err := security.HashPIN(pin)
	require.NoError(t, err)

	tests := []struct {
		name    string
		pin     string
		hash    string
		wantErr bool
	}{
		{
			name: "correct PIN",
			pin:  pin,
			hash: hash,
		},
		{
			name:    "incorrect PIN",
			pin:     "481517",
			hash:    hash,
			wantErr: true,
		},
		{
			name:    "invalid hash format",
			pin:     pin,
			hash:    "not-a-hash",
			wantErr: true,
		},
		{
			name:    "invalid algorithm",
			pin:     pin,
			hash:    "$sha256$v=19$salt$hash",
			wantErr: true,
		},
		{
			name:    "malformed parameters",
			pin:     pin,
			hash:    "$argon2id$v=19$m=invalid,t=2,p=2$c2FsdA$aGFzaA",
			wantErr: true,
		},
		{
			name:    "invalid salt base64",
			pin:     pin,
			hash:    "$argon2id$v=19$m=65536,t=2,p=2$!!!invalid!!!$aGFzaA",
			wantErr: true,
		},
		{
			name:    "invalid hash base64",
			pin:     pin,
			hash:    "$argon2id$v=19$m=65536,t=2,p=2$c2FsdA$!!!invalid!!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.VerifyPIN(tt.pin, tt.hash)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestGenerateSessionToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		token, err := security.GenerateSessionToken()

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, len(token), 30)
	})

	t.Run("URL safe", func(t *testing.T) {
		token, err := security.GenerateSessionToken()

		require.NoError(t, err)
		assert.NotContains(t, token, "+")
		assert.NotContains(t, token, "/")
	})
}

func TestGenerateSessionToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)

	for range 100 {
		token, err := security.GenerateSessionToken()

		require.NoError(t, err)
		assert.False(t, tokens[token], "generated duplicate token")

		tokens[token] = true
	}

	assert.Len(t, tokens, 100)
}

func TestHashAndVerifyPIN_Integration(t *testing.T) {
	validPINs := []string{
		"135792",
		"246802",
		"481516",
		"192837",
	}

	for _, pin := range validPINs {
		t.Run(pin, func(t *testing.T) {
			require.NoError(t, security.ValidatePIN(pin))

			hash, err := security.HashPIN(pin)
			require.NoError(t, err)

			assert.NoError(t, security.VerifyPIN(pin, hash))

			wrongPIN := pin[:5] + "9"
			assert.Error(t, security.VerifyPIN(wrongPIN, hash))
		})
	}
}
