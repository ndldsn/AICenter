package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// InitKey uses the dev fallback key when AICENTER_ENCRYPTION_KEY is unset.
	InitKey(nil)
	plaintext := "sk-proj-super-secret-key-123456"
	enc, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	// Ciphertext must never leak the plaintext.
	if strings.Contains(enc, plaintext) {
		t.Fatalf("ciphertext leaks plaintext: %s", enc)
	}
	dec, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if dec != plaintext {
		t.Fatalf("round-trip mismatch: got %q want %q", dec, plaintext)
	}
}

func TestDecryptInvalidBlobFails(t *testing.T) {
	InitKey(nil)
	if _, err := Decrypt("not-a-valid-hex!!"); err == nil {
		t.Fatal("expected decryption to fail on invalid hex")
	}
	if _, err := Decrypt(hexEncodeShort()); err == nil {
		t.Fatal("expected decryption to fail on too-short blob")
	}
}

func hexEncodeShort() string {
	// 4 bytes — shorter than a GCM nonce (12) + tag (16), guaranteed to fail.
	return "deadbeef"
}

func TestHint(t *testing.T) {
	h := Hint("sk-proj-abcdefghijklmnop")
	if !strings.HasPrefix(h, "sk-") {
		t.Fatalf("hint should preserve prefix, got %q", h)
	}
	if strings.Contains(h, "abcdefghijklmnop") {
		t.Fatalf("hint must not reveal the full secret: %q", h)
	}
	if Hint("short") != "-" {
		t.Fatalf("short key should yield '-'")
	}
}
