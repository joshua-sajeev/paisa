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

var ErrInvalidPIN = errors.New("invalid PIN")

// argon2 parameters - tune for your environment
const (
	argonTime    = 2
	argonMemory  = 64 * 1024
	argonThreads = 2
	argonKeyLen  = 32
	saltLen      = 16
)

// ValidatePIN enforces basic rules and rejects common weak PINs
func ValidatePIN(pin string) error {
	if m, _ := regexp.MatchString("^[0-9]{4,6}$", pin); !m {
		return fmt.Errorf("%w: PIN must be 4-6 digits", ErrInvalidPIN)
	}

	// reject obvious weak/repetitive/sequential PINs
	weak := map[string]bool{
		"0000": true, "1111": true, "2222": true, "3333": true, "4444": true,
		"5555": true, "6666": true, "7777": true, "8888": true, "9999": true,
		"0123": true, "1234": true, "2345": true, "3456": true, "4567": true,
		"5678": true, "6789": true,
	}
	if weak[pin] {
		return fmt.Errorf("%w: PIN is too common", ErrInvalidPIN)
	}

	// check sequential: either increasing or decreasing
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

	hash := argon2.IDKey([]byte(pin), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	phc := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime, argonThreads, b64Salt, b64Hash)
	return phc, nil
}

// VerifyPIN checks a PIN against a PHC-encoded Argon2id hash.
func VerifyPIN(pin, phc string) error {
	// phc format: $argon2id$v=19$m=65536,t=2,p=2$<salt>$<hash>
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return fmt.Errorf("invalid hash format")
	}

	params := parts[3]
	var m, t uint32
	var p uint8
	// params like m=65536,t=2,p=2
	_, err := fmt.Sscanf(params, "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return fmt.Errorf("invalid hash params: %w", err)
	}

	b64Salt := parts[4]
	b64Hash := parts[5]

	salt, err := base64.RawStdEncoding.DecodeString(b64Salt)
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(b64Hash)
	if err != nil {
		return fmt.Errorf("decode hash: %w", err)
	}

	computed := argon2.IDKey([]byte(pin), salt, t, m, p, uint32(len(expectedHash)))

	if len(computed) != len(expectedHash) {
		return errors.New("verification failed")
	}
	// constant time compare
	var diff byte
	for i := range computed {
		diff |= computed[i] ^ expectedHash[i]
	}
	if diff != 0 {
		return errors.New("verification failed")
	}
	return nil
}

// GenerateSessionToken returns a URL-safe random token of 32 bytes encoded.
func GenerateSessionToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
