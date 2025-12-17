package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/controlewise/backend/internal/config"
	"github.com/google/uuid"
)

type StorageService struct {
	cfg      config.StorageConfig
	s3Client *s3.Client
}

func NewStorageService(cfg config.StorageConfig) *StorageService {
	// Initialize S3 client if credentials are provided
	var s3Client *s3.Client
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.AWSRegion),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AWSAccessKeyID,
				cfg.AWSSecretAccessKey,
				"",
			)),
		)
		if err == nil {
			s3Client = s3.NewFromConfig(awsCfg)
		}
	}

	return &StorageService{
		cfg:      cfg,
		s3Client: s3Client,
	}
}

type UploadResult struct {
	FileName     string
	FileSize     int64
	MimeType     string
	URL          string
	ThumbnailURL string
}

func (s *StorageService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, orgID uuid.UUID) (*UploadResult, error) {
	// Validate file size
	if header.Size > s.cfg.MaxUploadSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	// Validate file type
	mimeType := header.Header.Get("Content-Type")
	if !s.isAllowedFileType(mimeType) {
		return nil, fmt.Errorf("file type not allowed")
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%s/%s%s", orgID.String(), uuid.New().String(), ext)

	// If S3 client is configured, upload to S3
	if s.s3Client != nil {
		return s.uploadToS3(ctx, file, uniqueFilename, mimeType, header.Size)
	}

	// Otherwise, return a placeholder (for local development)
	return &UploadResult{
		FileName: header.Filename,
		FileSize: header.Size,
		MimeType: mimeType,
		URL:      fmt.Sprintf("/uploads/%s", uniqueFilename),
	}, nil
}

func (s *StorageService) uploadToS3(ctx context.Context, file io.Reader, key, mimeType string, size int64) (*UploadResult, error) {
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.cfg.S3Bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentType:   aws.String(mimeType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.cfg.S3Bucket, s.cfg.AWSRegion, key)

	return &UploadResult{
		FileName: filepath.Base(key),
		FileSize: size,
		MimeType: mimeType,
		URL:      url,
	}, nil
}

func (s *StorageService) DeleteFile(ctx context.Context, url string) error {
	if s.s3Client == nil {
		return nil
	}

	// Extract key from URL
	// This is simplified - in production, you'd want more robust URL parsing
	key := filepath.Base(url)

	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.S3Bucket),
		Key:    aws.String(key),
	})

	return err
}

func (s *StorageService) GeneratePresignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	if s.s3Client == nil {
		return "", fmt.Errorf("S3 client not configured")
	}

	presignClient := s3.NewPresignClient(s.s3Client)
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.S3Bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}

func (s *StorageService) isAllowedFileType(mimeType string) bool {
	for _, allowed := range s.cfg.AllowedFileTypes {
		if allowed == mimeType {
			return true
		}
	}
	return false
}
