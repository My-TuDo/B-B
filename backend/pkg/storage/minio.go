package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/My-TuDo/B-B/backend/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client
var bucketName string
var endpoint string
var useSSL bool

func Init(cfg *config.Config) *minio.Client {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to init MinIO client: %v", err))
	}

	bucketName = cfg.MinioBucket
	endpoint = cfg.MinioEndpoint
	useSSL = cfg.MinioUseSSL

	exists, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		panic(fmt.Sprintf("failed to check MinIO bucket: %v", err))
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{}); err != nil {
			panic(fmt.Sprintf("failed to create MinIO bucket: %v", err))
		}
	}

	Client = client
	return client
}

func UploadVideo(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := Client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage.UploadVideo: %w", err)
	}
	return nil
}

func GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := Client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("storage.GetPresignedURL: %w", err)
	}
	return url.String(), nil
}

func DeleteVideo(ctx context.Context, objectName string) error {
	err := Client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage.DeleteVideo: %w", err)
	}
	return nil
}

func UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := Client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage.UploadFile: %w", err)
	}
	return nil
}

func GetObjectURL(objectName string) string {
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint, bucketName, objectName)
}
