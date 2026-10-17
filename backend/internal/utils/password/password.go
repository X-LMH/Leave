// Package password 提供独立的 Argon2id 密码哈希与验证，不处理业务密码规则。
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// 采用 OWASP 建议的 Argon2id 基线；内存单位为 KiB。
	memory      = 19 * 1024
	iterations  = 2
	parallelism = 1
	saltLength  = 16
	keyLength   = 32
)

// ErrInvalidHash 表示哈希格式、版本或参数不受支持。
var ErrInvalidHash = errors.New("password: invalid or unsupported Argon2id hash")

// Hash 使用随机盐生成 PHC 格式哈希，返回值可直接保存，无需单独存储盐。
// 不裁剪密码，也不校验业务层的长度或复杂度规则。
func Hash(plainPassword string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(plainPassword), salt, iterations, memory, parallelism, keyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// Verify 校验哈希与原始密码；不匹配返回 false, nil，非法哈希返回 ErrInvalidHash。
// 仅支持 Argon2id v19，不接受明文或其他算法作为兼容回退。
func Verify(hashedPassword, plainPassword string) (bool, error) {
	// 在拆分及解码前限制长度，避免异常输入造成过量分配。
	if len(hashedPassword) > 512 {
		return false, ErrInvalidHash
	}
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, ErrInvalidHash
	}
	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false, ErrInvalidHash
	}
	m, err := parseParameter(params[0], "m=")
	if err != nil {
		return false, err
	}
	t, err := parseParameter(params[1], "t=")
	if err != nil {
		return false, err
	}
	p, err := parseParameter(params[2], "p=")
	if err != nil {
		return false, err
	}
	// 参数来自存储内容，必须在运行 Argon2 前限制资源成本和非法零值。
	// 当前工具支持至多 64 MiB、10 次迭代、4 路并行；调整策略时需同时审视这些上限。
	if p < 1 || p > 4 || m < 8*p || m > 64*1024 || t < 1 || t > 10 {
		return false, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 64 || strings.ContainsAny(parts[4], "\r\n") {
		return false, ErrInvalidHash
	}
	key, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(key) < 16 || len(key) > 64 || strings.ContainsAny(parts[5], "\r\n") {
		return false, ErrInvalidHash
	}
	actual := argon2.IDKey([]byte(plainPassword), salt, t, m, uint8(p), uint32(len(key)))
	return subtle.ConstantTimeCompare(actual, key) == 1, nil
}

func parseParameter(value, prefix string) (uint32, error) {
	if !strings.HasPrefix(value, prefix) {
		return 0, ErrInvalidHash
	}
	digits := strings.TrimPrefix(value, prefix)
	if digits == "" || strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return 0, ErrInvalidHash
	}
	n, err := strconv.ParseUint(digits, 10, 32)
	if err != nil {
		return 0, ErrInvalidHash
	}
	return uint32(n), nil
}
