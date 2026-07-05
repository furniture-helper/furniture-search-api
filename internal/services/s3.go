package services

import (
	"context"
	"fmt"
	"furniture-search-api/internal/config"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct{}

func NewS3Service() *S3Service {
	return &S3Service{}
}

func (*S3Service) getObjectSignedUrl(ctx context.Context, bucket string, key string) (string, error) {
	region, err := config.GetS3Region()
	if err != nil {
		return "", fmt.Errorf("error while getting S3 region: %w", err)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	presignClient := s3.NewPresignClient(s3Client)

	presignedReq, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(5*time.Minute))

	if err != nil {
		log.Fatalf("Failed to generate presigned URL: %v", err)
	}

	return presignedReq.URL, nil
}

func (s *S3Service) GetCrawledPageUrl(ctx context.Context, key string) (string, error) {
	bucket, err := config.GetCrawlerStorageS3Bucket()
	if err != nil {
		return "", fmt.Errorf("error while getting crawler stroage S3 bucket: %w", err)
	}
	return s.getObjectSignedUrl(ctx, bucket, key)
}

func (s *S3Service) GetMinimizedPageUrl(ctx context.Context, key string) (string, error) {
	bucket, err := config.GetMinimizedPagesS3Bucket()
	if err != nil {
		return "", fmt.Errorf("error while getting minimized pages S3 bucket: %w", err)
	}

	return s.getObjectSignedUrl(ctx, bucket, key)
}
