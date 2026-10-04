// Package file 提供本地文件保存工具，不负责 HTTP 参数解析和业务校验。
package file

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var ErrInvalidPath = errors.New("invalid file path")

// Save 将文件保存到 root 下的业务目录，返回使用 / 分隔的相对文件路径。
// directory 和 extension 由服务端指定；调用方负责校验文件内容并限制大小。
// root 应使用受控的持久化目录，目录内不能放置指向外部的符号链接。
func Save(root, directory, extension string, content io.Reader) (string, error) {
	if root == "" || !validPath(directory) {
		return "", ErrInvalidPath
	}

	// 扩展名校验
	if len(extension) < 2 || extension[0] != '.' {
		return "", ErrInvalidPath
	}
	for _, character := range extension[1:] {
		isLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		if !isLetter && !isDigit {
			return "", ErrInvalidPath
		}
	}

	// 随机文件名字
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}

	// 拼接类如 key = "avatar/a8f13c92712e44bd8c2f62e23ac9e92b.jpg"
	key := path.Join(directory, hex.EncodeToString(randomBytes[:])+extension)
	filename := filepath.Join(root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return "", err
	}

	// 独占创建，避免意外覆盖已有文件。
	destination, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(destination, content)
	closeErr := destination.Close()
	if err := errors.Join(copyErr, closeErr); err != nil {
		// 写入失败时删除不完整文件，同时保留清理失败的信息。
		return "", errors.Join(err, os.Remove(filename))
	}
	return key, nil
}

// Delete 删除 Save 返回的相对路径；文件已不存在时视为清理成功。
func Delete(root, key string) error {
	if root == "" || !validPath(key) {
		return ErrInvalidPath
	}
	err := os.Remove(filepath.Join(root, filepath.FromSlash(key)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// AccessURL 将访问前缀（例如 /uploads 或 https://example.com/uploads）与文件路径拼接。
// 此函数只生成地址，文件是否公开以及如何提供访问由业务和路由配置决定。
func AccessURL(baseURL, key string) (string, error) {
	if !validPath(key) {
		return "", ErrInvalidPath
	}
	return url.JoinPath(baseURL, key)
}

func validPath(value string) bool {
	// 相对路径统一使用 /，并拒绝 Windows 路径语法及非规范路径段。
	if strings.ContainsAny(value, "\\:") {
		return false
	}
	for segment := range strings.SplitSeq(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
