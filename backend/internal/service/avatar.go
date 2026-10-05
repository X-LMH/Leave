package service

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/utils/file"
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"strings"

	"gorm.io/gorm"
)

var ErrorInvalidAvatar = errors.New("avatar must be a JPEG or PNG image within 4096 pixels per side")

func UploadAvatar(studentID string, content io.Reader) (*dto.AvatarUploadResponse, error) {
	data, extension, err := readAvatar(content)
	if err != nil {
		return nil, err
	}
	profile, err := mysql.GetProfileByStuID(studentID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrorProfileIncomplete
	}
	root := config.Cfg.Storage.RootDir
	baseURL := config.Cfg.Storage.BaseURL
	key, err := file.Save(root, "avatars", extension, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	avatarURL, err := file.AccessURL(baseURL, key)
	if err != nil {
		return nil, errors.Join(err, file.Delete(root, key))
	}

	if err := mysql.UpdateProfileAvatar(studentID, key); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = ErrorProfileIncomplete
		}
		return nil, errors.Join(err, file.Delete(root, key))
	}
	// 更新已成功，旧文件清理失败只记录日志，避免客户端重复上传。
	if err := deleteOldAvatar(root, profile.AvatarURL); err != nil {
		log.Printf("清理旧头像失败 student_id=%s: %v", studentID, err)
	}
	return &dto.AvatarUploadResponse{AvatarURL: avatarURL}, nil
}

func deleteOldAvatar(root, key string) error {
	// 默认头像是共享资源；空路径及其他业务目录也不应删除。
	if key == defaultAvatarPath || !strings.HasPrefix(key, "avatars/") {
		return nil
	}
	return file.Delete(root, key)
}

func readAvatar(content io.Reader) ([]byte, string, error) {
	// 调用方在解析上传请求时限制大小，此处只校验图片内容。
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", ErrorInvalidAvatar
	}
	// 根据图片内容判定格式，不信任客户端的文件名或 Content-Type。
	metadata, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") {
		return nil, "", ErrorInvalidAvatar
	}
	// 解码前限制尺寸，防止小文件解压后占用大量内存。
	if metadata.Width <= 0 || metadata.Height <= 0 || metadata.Width > 4096 || metadata.Height > 4096 {
		return nil, "", ErrorInvalidAvatar
	}
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		return nil, "", ErrorInvalidAvatar
	}
	extension := ".png"
	if format == "jpeg" {
		extension = ".jpg"
	}
	return data, extension, nil
}
