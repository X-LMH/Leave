package password

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestHashAndVerify(t *testing.T) {
	for _, plain := range []string{"123456", " 密码 with spaces ", strings.Repeat("长", 40), ""} {
		t.Run(plain, func(t *testing.T) {
			hash, err := Hash(plain)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") || len(hash) > 255 {
				t.Fatal("unexpected storage format")
			}
			if ok, err := Verify(hash, plain); err != nil || !ok {
				t.Fatalf("correct password rejected: %v", err)
			}
			if ok, err := Verify(hash, plain+"x"); err != nil || ok {
				t.Fatalf("incorrect password result: %v, %v", ok, err)
			}
			if strings.TrimSpace(plain) != plain {
				if ok, err := Verify(hash, strings.TrimSpace(plain)); err != nil || ok {
					t.Fatal("password whitespace was ignored")
				}
			}
			second, err := Hash(plain)
			if err != nil || second == hash {
				t.Fatalf("random salt not reflected in hash: %v", err)
			}
		})
	}
}

func TestVerifyStoredParameters(t *testing.T) {
	// 使用不同于默认值的参数，确保验证读取存储参数而非当前生成参数。
	salt := []byte("0123456789abcdef")
	key := argon2.IDKey([]byte("test-password"), salt, 1, 32*1024, 2, 16)
	hash := "$argon2id$v=19$m=32768,t=1,p=2$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key)
	if ok, err := Verify(hash, "test-password"); err != nil || !ok {
		t.Fatalf("stored parameters rejected: %v", err)
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	const valid = "$argon2id$v=19$m=19456,t=2,p=1$MDEyMzQ1Njc4OWFiY2RlZg$MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY"
	tests := []string{
		"", "plaintext", strings.Repeat("x", 513), valid + "$extra",
		strings.Replace(valid, "argon2id", "argon2i", 1),
		strings.Replace(valid, "v=19", "v=16", 1),
		strings.Replace(valid, "m=19456", "m=0", 1),
		strings.Replace(valid, "m=19456", "m=65537", 1),
		strings.Replace(valid, "m=19456", "m=4294967296", 1),
		strings.Replace(valid, "m=19456", "m=+19456", 1),
		strings.Replace(valid, "t=2", "t=0", 1),
		strings.Replace(valid, "t=2", "t=11", 1),
		strings.Replace(valid, "t=2", "t=-1", 1),
		strings.Replace(valid, "p=1", "p=0", 1),
		strings.Replace(valid, "p=1", "p=5", 1),
		strings.Replace(valid, "t=2,p=1", "p=1,t=2", 1),
		strings.Replace(valid, "t=2,p=1", "t=2", 1),
		strings.Replace(valid, "t=2", "t=", 1),
		strings.Replace(valid, "MDEyMzQ1Njc4OWFiY2RlZg", "bad!", 1),
		strings.Replace(valid, "MDEyMzQ1Njc4OWFiY2RlZg", "YQ", 1),
		valid[:strings.LastIndex(valid, "$")+1] + "YQ",
		valid[:strings.LastIndex(valid, "$")+1] + "bad!",
		valid[:strings.LastIndex(valid, "$")+1] + base64.RawStdEncoding.EncodeToString(make([]byte, 65)),
		strings.Replace(valid, "MDEyMzQ1Njc4OWFiY2RlZg", base64.RawStdEncoding.EncodeToString(make([]byte, 65)), 1),
		valid + "\n",
	}
	for _, hash := range tests {
		if ok, err := Verify(hash, "password"); ok || !errors.Is(err, ErrInvalidHash) {
			t.Errorf("invalid hash accepted: %q: %v, %v", hash, ok, err)
		}
	}
}
