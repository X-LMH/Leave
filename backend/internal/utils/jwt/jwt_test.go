package jwt

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseToken(t *testing.T) {
	tokens := NewManager("test-secret", 1)

	token, err := tokens.GenerateToken("202600010001", "student")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := tokens.ParseToken(token)
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
	tokens := NewManager("test-secret", 1)

	token, err := tokens.GenerateToken("202600010001", "student")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	// 修改实际签名字节，避免只改变 Base64URL 末尾未使用的位。
	signature[0] ^= 1
	parts[2] = base64.RawURLEncoding.EncodeToString(signature)
	if _, err := tokens.ParseToken(strings.Join(parts, ".")); err == nil {
		t.Fatal("expected tampered token to be rejected")
	}

	wrongMethodToken := jwtv5.NewWithClaims(jwtv5.SigningMethodHS384, &Claims{
		StudentID: "202600010001",
		Role:      "student",
	})
	wrongMethodTokenString, err := wrongMethodToken.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tokens.ParseToken(wrongMethodTokenString); err == nil {
		t.Fatal("expected token with wrong signing method to be rejected")
	}
}

func TestManagersKeepIndependentSigningAndExpiration(t *testing.T) {
	short := NewManager("short-session-secret", 1)
	long := NewManager("long-session-secret", 7)
	for _, tc := range []struct {
		issuer *Manager
		other  *Manager
		days   int
	}{
		{issuer: short, other: long, days: 1},
		{issuer: long, other: short, days: 7},
	} {
		token, err := tc.issuer.GenerateToken("student", "student")
		if err != nil {
			t.Fatal(err)
		}
		claims, err := tc.issuer.ParseToken(token)
		if err != nil {
			t.Fatal(err)
		}
		want := time.Now().AddDate(0, 0, tc.days)
		if delta := claims.ExpiresAt.Sub(want); delta < -5*time.Second || delta > 5*time.Second {
			t.Fatalf("unexpected expiration for %d-day manager: %v", tc.days, claims.ExpiresAt)
		}
		if _, err := tc.other.ParseToken(token); err == nil {
			t.Fatal("another manager accepted a token signed with a different key")
		}
	}
}
