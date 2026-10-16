package file

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

var ErrInvalidImage = errors.New("invalid JPEG or PNG image")

// ReadImage 根据实际内容校验 JPEG/PNG、尺寸和完整解码，返回原始数据与扩展名。
// 调用方负责通过受限 Reader 控制读取大小；maxDimension 指定单边最大像素数。
func ReadImage(content io.Reader, maxDimension int) ([]byte, string, error) {
	// 调用方在解析上传请求时限制大小，此处只校验图片内容。
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", ErrInvalidImage
	}
	// 根据图片内容判定格式，不信任客户端的文件名或 Content-Type。
	metadata, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") {
		return nil, "", ErrInvalidImage
	}
	// 解码前限制尺寸，防止小文件解压后占用大量内存。
	if metadata.Width <= 0 || metadata.Height <= 0 || metadata.Width > maxDimension || metadata.Height > maxDimension {
		return nil, "", ErrInvalidImage
	}
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		return nil, "", ErrInvalidImage
	}
	extension := ".png"
	if format == "jpeg" {
		extension = ".jpg"
	}
	return data, extension, nil
}
