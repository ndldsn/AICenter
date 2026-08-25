package auth

import (
	"testing"
	"time"
)

const testSecret = "test-jwt-secret-key-that-is-long-enough"

// --- GenerateTokenPair ---

func TestGenerateTokenPair_Success(t *testing.T) {
	access, refresh, err := GenerateTokenPair("uid-1", "alice", "admin", testSecret, 5*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}
	if access == "" {
		t.Fatal("access token must not be empty")
	}
	if refresh == "" {
		t.Fatal("refresh token must not be empty")
	}
	if access == refresh {
		t.Fatal("access and refresh tokens must differ")
	}
}

func TestGenerateTokenPair_EmptySecret_ProducesToken(t *testing.T) {
	// JWT library allows empty secret (used in testing); we verify it still produces tokens.
	access, refresh, err := GenerateTokenPair("uid-1", "alice", "admin", "", 5*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("empty secret should not error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected non-empty tokens even with empty secret")
	}
}

func TestGenerateTokenPair_ContainsCorrectClaims(t *testing.T) {
	access, _, err := GenerateTokenPair("uid-42", "bob", "operator", testSecret, 5*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}
	claims, err := ValidateAccessToken(access, testSecret)
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if claims.UserID != "uid-42" {
		t.Fatalf("expected UserID uid-42, got %s", claims.UserID)
	}
	if claims.Username != "bob" {
		t.Fatalf("expected Username bob, got %s", claims.Username)
	}
	if claims.Role != "operator" {
		t.Fatalf("expected Role operator, got %s", claims.Role)
	}
	if claims.Issuer != "aicenter" {
		t.Fatalf("expected Issuer aicenter, got %s", claims.Issuer)
	}
}

// --- ValidateAccessToken ---

func TestValidateAccessToken_ValidToken(t *testing.T) {
	access, _, _ := GenerateTokenPair("uid-1", "alice", "admin", testSecret, 5*time.Minute, 7*24*time.Hour)
	claims, err := ValidateAccessToken(access, testSecret)
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if claims.UserID != "uid-1" {
		t.Fatalf("expected UserID uid-1, got %s", claims.UserID)
	}
}

func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	access, _, _ := GenerateTokenPair("uid-1", "alice", "admin", testSecret, -1*time.Second, 7*24*time.Hour)
	_, err := ValidateAccessToken(access, testSecret)
	if err == nil {
		t.Fatal("expected error for expired access token")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	access, _, _ := GenerateTokenPair("uid-1", "alice", "admin", testSecret, 5*time.Minute, 7*24*time.Hour)
	_, err := ValidateAccessToken(access, "wrong-secret-key")
	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}

func TestValidateAccessToken_MalformedToken(t *testing.T) {
	_, err := ValidateAccessToken("not-a-jwt", testSecret)
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestValidateAccessToken_EmptyToken(t *testing.T) {
	_, err := ValidateAccessToken("", testSecret)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

// --- ValidateRefreshToken ---

func TestValidateRefreshToken_ValidToken(t *testing.T) {
	_, refresh, _ := GenerateTokenPair("uid-99", "carol", "viewer", testSecret, 5*time.Minute, 7*24*time.Hour)
	subject, err := ValidateRefreshToken(refresh, testSecret)
	if err != nil {
		t.Fatalf("ValidateRefreshToken: %v", err)
	}
	if subject != "uid-99" {
		t.Fatalf("expected subject uid-99, got %s", subject)
	}
}

func TestValidateRefreshToken_ExpiredToken(t *testing.T) {
	_, refresh, _ := GenerateTokenPair("uid-1", "alice", "admin", testSecret, 5*time.Minute, -1*time.Second)
	_, err := ValidateRefreshToken(refresh, testSecret)
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
}

func TestValidateRefreshToken_WrongSecret(t *testing.T) {
	_, refresh, _ := GenerateTokenPair("uid-1", "alice", "admin", testSecret, 5*time.Minute, 7*24*time.Hour)
	_, err := ValidateRefreshToken(refresh, "wrong-secret")
	if err == nil {
		t.Fatal("expected error when validating refresh token with wrong secret")
	}
}

func TestValidateRefreshToken_MissingSubject(t *testing.T) {
	// A refresh token without a subject claim should be rejected.
	// We craft this by generating an access token (which has UserID not Subject)
	// and trying to validate it as a refresh token.
	access, _, _ := GenerateTokenPair("uid-1", "alice", "admin", testSecret, 5*time.Minute, 7*24*time.Hour)
	_, err := ValidateRefreshToken(access, testSecret)
	if err == nil {
		t.Fatal("expected error for refresh token missing subject claim")
	}
}

// --- HashPassword / CheckPasswordHash ---

func TestHashPassword_AndCheck(t *testing.T) {
	hash, err := HashPassword("MySecurePass123!")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !CheckPasswordHash("MySecurePass123!", hash) {
		t.Fatal("CheckPasswordHash should return true for correct password")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("CorrectPass")
	if CheckPasswordHash("WrongPass", hash) {
		t.Fatal("CheckPasswordHash should return false for wrong password")
	}
}

func TestCheckPasswordHash_EmptyPassword(t *testing.T) {
	hash, _ := HashPassword("NonEmpty")
	if CheckPasswordHash("", hash) {
		t.Fatal("CheckPasswordHash should return false for empty password")
	}
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	hash1, _ := HashPassword("SamePass123!")
	hash2, _ := HashPassword("SamePass123!")
	if hash1 == hash2 {
		t.Fatal("bcrypt should produce different hashes for the same password (different salts)")
	}
}