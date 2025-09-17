package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
)

func UploadImageToMinio(ctx context.Context, minioCli *minio.Client, bucket, localPath, minioPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local file failed: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file info failed: %w", err)
	}

	_, err = minioCli.PutObject(
		ctx,
		bucket,
		strings.TrimPrefix(minioPath, "/"), // 移除开头的斜杠
		file,
		fileInfo.Size(),
		minio.PutObjectOptions{
			ContentType: "image/jpeg",
		},
	)
	if err != nil {
		return fmt.Errorf("put object to minio failed: %w", err)
	}

	return nil
}
