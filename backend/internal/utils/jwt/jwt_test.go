package jwt

import (
	"backend/internal/config"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseToken(t *testing.T) {
	previous := config.Cfg
	config.Cfg.JWT.Secret = "test-secret"
	config.Cfg.JWT.ExpirationDays = 1
	t.Cleanup(func() { config.Cfg = previous })

	token, err := GenerateToken("202600010001", "student")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.StudentID != "202600010001" || claims.Role != "student" || claims.Issuer != "LMH" {
		t.Fatalf("claims = %#v", claims)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(time.Now()) {
		t.Fatalf("expiration claim = %#v", claims.ExpiresAt)
	}
}

func TestParseTokenRejectsTamperedAndWrongAlgorithmTokens(t *testing.T) {
	previous := config.Cfg
	config.Cfg.JWT.Secret = "test-secret"
	config.Cfg.JWT.ExpirationDays = 1
	t.Cleanup(func() { config.Cfg = previous })

	token, err := GenerateToken("202600010001", "student")
	if err != nil {
		t.Fatal(err)
	}
	parts := []byte(token)
	parts[len(parts)-1] ^= 1
	if _, err := ParseToken(string(parts)); err == nil {
		t.Fatal("expected tampered token to be rejected")
	}

	wrongMethodToken := jwtv5.NewWithClaims(jwtv5.SigningMethodHS384, &Claims{
		StudentID: "202600010001",
		Role:      "student",
	})
	wrongMethodTokenString, err := wrongMethodToken.SignedString(config.JWTSecret())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken(wrongMethodTokenString); err == nil {
		t.Fatal("expected token with wrong signing method to be rejected")
	}
}
