// Package storage 提供 MinIO 对象存储的初始化与文件操作。
// 支持视频上传、预签名 URL 生成、文件删除和公开 URL 构建。
// 启动时自动检查并创建存储桶，连接或创建桶失败会直接 panic。
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

// 包级变量，由 Init 在服务启动时设置，供后续文件操作使用。
var Client *minio.Client  // MinIO 客户端实例（全局复用）
var bucketName string     // 存储桶名称
var endpoint string       // 内网端点地址
var publicEndpoint string // 公网端点地址（用于生成公开 URL）
var useSSL bool           // 是否启用 SSL 加密连接

// Init 初始化 MinIO 客户端，确保存储桶存在，设置包级变量。
// 连接失败或创建桶失败会直接 panic（对象存储是核心依赖，不可降级）。
// cfg 为应用配置，从中读取 MinIO 相关参数。
func Init(cfg *config.Config) *minio.Client {
	// 创建 MinIO 客户端
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to init MinIO client: %v", err))
	}

	// 设置包级变量（供其他函数使用）
	bucketName = cfg.MinioBucket
	endpoint = cfg.MinioEndpoint
	publicEndpoint = cfg.MinioPublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.MinioEndpoint // 公网端点未配置时回退到内网端点
	}
	useSSL = cfg.MinioUseSSL

	// 检查存储桶是否存在，不存在则创建
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

// UploadVideo 上传视频文件到 MinIO。
// objectName 为存储路径（如 "videos/xxx.mp4"）；
// reader 为文件内容流；
// size 为文件字节数；
// contentType 为 MIME 类型（如 "video/mp4"）。
func UploadVideo(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := Client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage.UploadVideo: %w", err)
	}
	return nil
}

// GetPresignedURL 生成文件的预签名下载 URL。
// objectName 为存储路径；expiry 为 URL 有效期。
// 返回可公开访问的临时下载链接。
func GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := Client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("storage.GetPresignedURL: %w", err)
	}
	return url.String(), nil
}

// DeleteVideo 从 MinIO 删除指定对象。
// objectName 为要删除的存储路径。
func DeleteVideo(ctx context.Context, objectName string) error {
	err := Client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage.DeleteVideo: %w", err)
	}
	return nil
}

// UploadFile 上传通用文件到 MinIO（封面、头像等非视频文件）。
// 参数含义与 UploadVideo 相同，语义上用于区分视频和普通文件。
func UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := Client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage.UploadFile: %w", err)
	}
	return nil
}

// GetObjectURL 构建对象的公开访问 URL。
// 根据 useSSL 自动选择 http 或 https 协议；
// 使用 publicEndpoint 作为域名（支持 CDN 等场景）。
func GetObjectURL(objectName string) string {
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, publicEndpoint, bucketName, objectName)
}
