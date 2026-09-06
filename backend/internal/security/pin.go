// Package security provides PIN hashing, verification, and secure token generation.
package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ErrInvalidPIN indicates that a PIN does not satisfy the required format
// or strength rules.
var ErrInvalidPIN = errors.New("invalid PIN")

const (
	argonTime    = 2
	argonMemory  = 64 * 1024
	argonThreads = 2
	argonKeyLen  = 32
	saltLen      = 16
)

// ValidatePIN enforces a 6-digit PIN and rejects common weak PINs.
func ValidatePIN(pin string) error {
	if m, _ := regexp.MatchString(`^[0-9]{6}$`, pin); !m {
		return fmt.Errorf("%w: PIN must be 6 digits", ErrInvalidPIN)
	}

	weak := map[string]bool{
		"000000": true,
		"111111": true,
		"222222": true,
		"333333": true,
		"444444": true,
		"555555": true,
		"666666": true,
		"777777": true,
		"888888": true,
		"999999": true,
		"012345": true,
		"123456": true,
		"234567": true,
		"345678": true,
		"456789": true,
		"987654": true,
		"876543": true,
		"765432": true,
		"654321": true,
		"543210": true,
	}

	if weak[pin] {
		return fmt.Errorf("%w: PIN is too common", ErrInvalidPIN)
	}

	isSeqInc := true
	isSeqDec := true

	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]+1 {
			isSeqInc = false
		}
		if pin[i] != pin[i-1]-1 {
			isSeqDec = false
		}
	}

	if isSeqInc || isSeqDec {
		return fmt.Errorf("%w: PIN cannot be sequential", ErrInvalidPIN)
	}

	return nil
}

// HashPIN returns an Argon2id PHC encoded string for storage.
func HashPIN(pin string) (string, error) {
	if err := ValidatePIN(pin); err != nil {
		return "", err
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(pin),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		b64Salt,
		b64Hash,
	), nil
}

// VerifyPIN verifies a PIN against a stored Argon2id PHC hash.
func VerifyPIN(pin, phc string) error {
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return fmt.Errorf("invalid hash format")
	}

	params := parts[3]

	var m, t uint32
	var p uint8

	_, err := fmt.Sscanf(params, "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return fmt.Errorf("invalid hash params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("decode hash: %w", err)
	}

	computed := argon2.IDKey(
		[]byte(pin),
		salt,
		t,
		m,
		p,
		uint32(len(expectedHash)),
	)

	if len(computed) != len(expectedHash) {
		return errors.New("verification failed")
	}

	var diff byte
	for i := range computed {
		diff |= computed[i] ^ expectedHash[i]
	}

	if diff != 0 {
		return errors.New("verification failed")
	}

	return nil
}

// GenerateSessionToken returns a URL-safe random token.
func GenerateSessionToken() (string, error) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}
