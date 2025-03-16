package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var instance *minio.Client

// Init initializes the MinIO client
func Init() {
	var endpoint string = os.Getenv("MINIO_ENDPOINT")
	var accessKeyID string = os.Getenv("MINIO_ACCESS_KEY")
	var secretAccessKey string = os.Getenv("MINIO_SECRET_KEY")
	var bucket string = os.Getenv("DEFAULT_BUCKET")

	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Fatalln("Failed to initialize MinIO client:", err)
	}

	instance = minioClient
	log.Println("MinIO client initialized successfully.")

	CreateBucket(bucket)
}

// CreateBucket ensures the specified bucket exists
func CreateBucket(bucket string) {
	ctx := context.Background()

	exists, err := instance.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}

	if !exists {
		err = instance.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		log.Printf("Bucket '%s' created successfully.\n", bucket)
	} else {
		log.Printf("Bucket '%s' already exists.\n", bucket)
	}
}

// UploadFile uploads a file to MinIO
func UploadFile(bucket, filePath, objectName string) {
	ctx := context.Background()

	uploadInfo, err := instance.FPutObject(ctx, bucket, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	log.Printf("File '%s' uploaded successfully to bucket '%s'.\n", objectName, bucket)
	fmt.Printf("Upload Details: %+v\n", uploadInfo)
}

// Uploads logs as byte stream to MinIO
func UploadLogsToMinIO(bucket, objectName string, logData *bytes.Buffer) error {
	ctx := context.Background()

	// Upload logs as an object to MinIO
	_, err := instance.PutObject(ctx, bucket, objectName, logData, int64(logData.Len()), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		return fmt.Errorf("failed to upload logs to MinIO: %w", err)
	}

	log.Printf("Logs uploaded successfully to MinIO bucket '%s' as '%s'.\n", bucket, objectName)
	return nil
}
