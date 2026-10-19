package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/service"
	"backend/internal/utils/file"
	"bytes"
	"errors"
	"io"
	"log"
	"strings"

	"gorm.io/gorm"
)

type AvatarService struct {
	dao     *mysql.DAO
	storage config.StorageConfig
}

func NewAvatarService(dao *mysql.DAO, storage config.StorageConfig) *AvatarService {
	return &AvatarService{
		dao:     dao,
		storage: storage,
	}
}

var ErrorInvalidAvatar = errors.New("avatar must be a JPEG or PNG image within 4096 pixels per side")

func (a *AvatarService) UploadAvatar(studentID string, content io.Reader) (*dto.AvatarUploadResponse, error) {
	data, extension, err := readAvatar(content)
	if err != nil {
		return nil, err
	}
	profile, err := a.dao.GetProfileByStuID(studentID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrorProfileIncomplete
	}
	root := a.storage.RootDir
	baseURL := a.storage.BaseURL
	keys, err := file.Save(root, "avatars", file.Input{
		Extension: extension,
		Content:   bytes.NewReader(data),
	})
	if err != nil {
		return nil, err
	}
	key := keys[0]
	avatarURL, err := file.AccessURL(baseURL, key)
	if err != nil {
		return nil, errors.Join(err, file.Delete(root, key))
	}

	if err := a.dao.UpdateProfileAvatar(studentID, key); err != nil {
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
	if key == service.DefaultAvatarPath || !strings.HasPrefix(key, "avatars/") {
		return nil
	}
	return file.Delete(root, key)
}

func readAvatar(content io.Reader) ([]byte, string, error) {
	data, extension, err := file.ReadImage(content, 4096)
	if errors.Is(err, file.ErrInvalidImage) {
		return nil, "", ErrorInvalidAvatar
	}
	return data, extension, err
}
