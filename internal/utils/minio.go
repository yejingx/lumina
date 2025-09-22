package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
)

func UploadFileToMinio(ctx context.Context, minioCli *minio.Client, bucket, localPath, minioPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local file failed: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file info failed: %w", err)
	}

	// Get file extension and determine content type
	lastDotIndex := strings.LastIndex(localPath, ".")
	ext := ""
	if lastDotIndex != -1 {
		ext = strings.ToLower(localPath[lastDotIndex+1:])
	}
	contentType := "application/octet-stream" // default content type

	switch ext {
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	case "gif":
		contentType = "image/gif"
	case "webp":
		contentType = "image/webp"
	case "pdf":
		contentType = "application/pdf"
	case "txt":
		contentType = "text/plain"
	case "html", "htm":
		contentType = "text/html"
	case "json":
		contentType = "application/json"
	case "xml":
		contentType = "application/xml"
	case "mp4":
		contentType = "video/mp4"
	case "avi":
		contentType = "video/avi"
	case "mov":
		contentType = "video/quicktime"
	}

	_, err = minioCli.PutObject(
		ctx,
		bucket,
		strings.TrimPrefix(minioPath, "/"), // 移除开头的斜杠
		file,
		fileInfo.Size(),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("put object to minio failed: %w", err)
	}

	return nil
}
