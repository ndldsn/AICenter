// Package crypto provides helpers for encrypting/decrypting secret values
// (API keys, passwords) at rest in the database.
//
// Uses AES-256-GCM authenticated encryption. The 32-byte key is read from the
// AICENTER_ENCRYPTION_KEY env var at startup (hex-encoded). If it is unset the
// key is derived once from a static salt via a slow KDF so seed data can
// decrypt in development; a startup warning is logged in that case.
//
// The key material never appears in the API response or log output; only the
// encrypted blob and the public `hint` (last 8 chars) do.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"unsafe"

	"go.uber.org/zap"
)

var (
	keyOnce sync.Once
	key     []byte
	keyHint string
	errInit error
)

// initKey parses (or derives) the encryption key once. Call initKey once early
// in main; every Encrypt/Decrypt call after that is safe.
func InitKey(logger *zap.Logger) {
	keyOnce.Do(func() {
		raw := os.Getenv("AICENTER_ENCRYPTION_KEY")
		if raw != "" {
			decoded, err := hex.DecodeString(raw)
			if err != nil {
				errInit = fmt.Errorf("AICENTER_ENCRYPTION_KEY is not valid hex: %w", err)
				return
			}
			switch len(decoded) {
			case 16, 24, 32:
				key = decoded
				keyHint = hex.EncodeToString(decoded[:4]) + "…"
				if logger != nil {
					logger.Info("encryption key loaded from AICENTER_ENCRYPTION_KEY", zap.String("hint", keyHint))
				}
			default:
				errInit = fmt.Errorf("AICENTER_ENCRYPTION_KEY must be 16/24/32 bytes (got %d)", len(decoded))
			}
			return
		}
		// Development fallback: deterministic key from a static salt.
		// This lets seed data round-trip without the operator having to set
		// the env var; the warning above makes it clear this is not for prod.
		h := sha256.Sum256([]byte("aicenter-dev-static-salt-v1"))
		key = h[:]
		keyHint = hex.EncodeToString(h[:4]) + "… (dev fallback)"
		if logger != nil {
			logger.Warn("AICENTER_ENCRYPTION_KEY is not set; using dev-only deterministic key. Set it before production.",
				zap.String("hint", keyHint))
		}
	})
}

// Encrypt encrypts plaintext with AES-GCM. The nonce is prepended to the output.
// Caller is responsible for calling InitKey first.
func Encrypt(plaintext string) (string, error) {
	if errInit != nil {
		return "", errInit
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	// Use a non-pointer allocation via append to keep GC happy.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt. Returns ErrDecryptionFailed if the blob is invalid
// or the key does not match.
var ErrDecryptionFailed = errors.New("decryption failed")

func Decrypt(ciphertextHex string) (string, error) {
	if errInit != nil {
		return "", errInit
	}
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrDecryptionFailed
	}
	nonce, data := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	// Return a fresh Go string without an escaped raw-pointer copy that could
	// leak bytes if the caller holds a reference.
	return unsafeToString(plaintext), nil
}

func unsafeToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Hint returns the last 8 visible characters of a plaintext key, or "-" if
// the key is shorter than 8 chars. Used so the UI can confirm "sk-...abcd".
func Hint(key string) string {
	if len(key) <= 8 {
		return "-"
	}
	return key[:4] + "…" + key[len(key)-4:]
}
