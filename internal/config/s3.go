package config

import (
	"errors"
	"os"
)

func GetS3Region() (string, error) {
	region := os.Getenv("S3_REGION")
	if region == "" {
		return "", errors.New("S3_REGION environment variable not set")
	}

	return region, nil
}

func GetCrawlerStorageS3Bucket() (string, error) {
	bucket := os.Getenv("CRAWLER_STORAGE_S3_BUCKET")
	if bucket == "" {
		return "", errors.New("CRAWLER_STORAGE_S3_BUCKET environment variable not set")
	}

	return bucket, nil
}

func GetMinimizedPagesS3Bucket() (string, error) {
	bucket := os.Getenv("MINIMIZED_PAGES_S3_BUCKET")
	if bucket == "" {
		return "", errors.New("MINIMIZED_PAGES_S3_BUCKET environment variable not set")
	}

	return bucket, nil
}
