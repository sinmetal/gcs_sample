package encryption_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/sinmetal/gcs_sample/encryption"
	"google.golang.org/api/cloudkms/v1"
)

func TestCSEKService_Download(t *testing.T) {
	ctx := context.Background()

	s := newCSEKService(ctx, t)

	keyName := os.Getenv("CLOUDKMS_KEY")
	bucketName := os.Getenv("BUCKET_NAME")
	object := uuid.New().String()
	encryptionKey, err := encryption.GenerateEncryptionKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("keyName=%s,bucket=%s,object=%s\n", keyName, bucketName, object)

	uploadText := []byte("Hello World")
	size, err := s.Upload(ctx, keyName, bucketName, object, encryptionKey, uploadText)
	if err != nil {
		t.Fatal(err)
	}
	if size < 1 {
		t.Fatal("Upload size is Zero")
	}

	got, _, err := s.Download(ctx, keyName, bucketName, object)
	if err != nil {
		t.Fatal(err)
	}

	if e, g := uploadText, got; bytes.Compare(e, g) != 0 {
		t.Errorf("want %s but got %s", string(e), string(g))
	}
}

func TestCSEKService_Copy(t *testing.T) {
	ctx := context.Background()

	s := newCSEKService(ctx, t)

	keyName := os.Getenv("CLOUDKMS_KEY")
	bucketName := os.Getenv("BUCKET_NAME")
	object := uuid.New().String()
	encryptionKey, err := encryption.GenerateEncryptionKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("keyName=%s,bucket=%s,object=%s\n", keyName, bucketName, object)

	uploadText := []byte("Hello World")
	size, err := s.Upload(ctx, keyName, bucketName, object, encryptionKey, uploadText)
	if err != nil {
		t.Fatal(err)
	}
	if size < 1 {
		t.Fatal("Upload size is Zero")
	}

	dstBucketName := fmt.Sprintf("%s-encrypt", bucketName)
	if err := s.Copy(ctx, dstBucketName, bucketName, object, keyName); err != nil {
		t.Fatal(err)
	}
}

func newCSEKService(ctx context.Context, t *testing.T) *encryption.CSEKService {
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	kms, err := cloudkms.NewService(ctx)
	if err != nil {
		t.Fatal(err)
	}
	s, err := encryption.NewCSEKService(ctx, gcs, kms)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
