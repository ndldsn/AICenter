package config

import "testing"

func TestLoad_RejectsMissingJWTSecret(t *testing.T) {
	// Ensure JWT_SECRET is not present, forcing Load() to use its validation path.
	t.Setenv("JWT_SECRET", "")

	cfg, err := Load()
	if err == nil {
		t.Fatalf("expected error when JWT_SECRET is unset, got nil; cfg=%+v", cfg)
	}
}

func TestLoad_AcceptsProvidedJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "review-verification-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error with JWT_SECRET set: %v", err)
	}
	if cfg == nil || cfg.Auth.Secret != "review-verification-secret" {
		t.Fatalf("expected Auth.Secret to be set from env, got %+v", cfg)
	}
}