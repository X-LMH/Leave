package app

import (
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppVersionHandler struct {
	logger     *zap.Logger
	appVersion *appservice.AppVersionService
	packageDir string
}

func NewAppVersionHandler(appVersion *appservice.AppVersionService, packageDir string, logger *zap.Logger) *AppVersionHandler {
	return &AppVersionHandler{
		logger:     logger,
		appVersion: appVersion,
		packageDir: packageDir,
	}
}

// GetCurrentAppVersionHandler returns the current published version for a platform.
func (a *AppVersionHandler) GetCurrentAppVersionHandler(c *gin.Context) {
	platform := strings.TrimSpace(c.Query("platform"))
	if platform == "" {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	versionCode, _ := strconv.Atoi(c.Query("version_code"))
	version, err := a.appVersion.GetCurrentAppVersion(platform, versionCode)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, version)
}

// DownloadCurrentAppHandler serves only the current published package for a platform.
func (a *AppVersionHandler) DownloadCurrentAppHandler(c *gin.Context) {
	platform := strings.TrimSpace(c.Query("platform"))
	if platform == "" {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	version, err := a.appVersion.GetCurrentAppPackage(platform)
	if err != nil {
		if errors.Is(err, appservice.ErrAppVersionNotFound) {
			response.Error(c, response.CodeFileNotFound)
			return
		}
		response.Error(c, response.CodeServerBusy)
		return
	}

	file, info, err := openAPK(a.packageDir, version.PackageFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			response.Error(c, response.CodeFileNotFound)
			return
		}
		response.Error(c, response.CodeServerBusy)
		return
	}
	defer file.Close()

	serveAPK(c, file, info)
}

func openAPK(packageDir, packageFile string) (*os.File, os.FileInfo, error) {
	if !isSafeAPKFileName(packageFile) {
		return nil, nil, fmt.Errorf("invalid APK package file name")
	}

	file, err := os.Open(filepath.Join(packageDir, packageFile))
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, nil, fmt.Errorf("APK package is not a regular file")
	}
	return file, info, nil
}

func isSafeAPKFileName(name string) bool {
	return name != "" &&
		filepath.Base(name) == name &&
		!filepath.IsAbs(name) &&
		!strings.ContainsAny(name, `\\/`) &&
		strings.HasSuffix(strings.ToLower(name), ".apk")
}

func serveAPK(c *gin.Context, file *os.File, info os.FileInfo) {
	filename := filepath.Base(info.Name())
	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, filename, info.ModTime(), file)
}
