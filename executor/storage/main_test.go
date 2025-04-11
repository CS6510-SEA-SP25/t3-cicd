package storage

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

func setupTestMinIO() *minio.Client {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}
	return client
}

func cleanupBucket(bucket string) {
	ctx := context.Background()
	objects := instance.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true})
	for object := range objects {
		instance.RemoveObject(ctx, bucket, object.Key, minio.RemoveObjectOptions{})
	}
	instance.RemoveBucket(ctx, bucket)
}

func TestInit(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("MINIO_ENDPOINT", "localhost:9000")
	os.Setenv("MINIO_ACCESS_KEY", "minioadmin")
	os.Setenv("MINIO_SECRET_KEY", "minioadmin")
	os.Setenv("DEFAULT_BUCKET", "test-bucket")

	// Call Init function
	Init()

	// Verify that instance is initialized
	assert.NotNil(t, instance)
}

func TestCreateBucket(t *testing.T) {
	instance = setupTestMinIO()
	bucket := "test-bucket"

	CreateBucket(bucket)

	exists, err := instance.BucketExists(context.Background(), bucket)
	assert.NoError(t, err)
	assert.True(t, exists)

	cleanupBucket(bucket)
}

func TestCreateBucket_Error(t *testing.T) {
	instance = setupTestMinIO()
	bucket := ""
	err := instance.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
	assert.Error(t, err)
}

func TestUploadLogsToMinIO(t *testing.T) {
	instance = setupTestMinIO()
	bucket := "test-bucket"
	CreateBucket(bucket)

	logData := bytes.NewBufferString("Test log data")
	objectName := "logs.txt"

	err := UploadLogsToMinIO(bucket, objectName, logData)
	assert.NoError(t, err)

	// Verify object exists
	_, err = instance.StatObject(context.Background(), bucket, objectName, minio.StatObjectOptions{})
	assert.NoError(t, err)

	cleanupBucket(bucket)
}
