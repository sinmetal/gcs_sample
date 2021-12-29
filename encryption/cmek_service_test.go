package encryption_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/sinmetal/gcs_sample/encryption"
)

func TestCMEKService_UploadWithKey(t *testing.T) {
	ctx := context.Background()

	keyName := os.Getenv("CLOUDKMS_KEY")
	bucketName := os.Getenv("BUCKET_NAME")
	object := uuid.New().String()
	t.Logf("keyName=%s,bucket=%s,object=%s\n", keyName, bucketName, object)

	uploadText := []byte("Hello World")

	s := newCMEKService(ctx, t)
	size, err := s.UploadWithKey(ctx, keyName, bucketName, object, uploadText)
	if size < 1 {
		t.Fatal("Upload size is Zero")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func newCMEKService(ctx context.Context, t *testing.T) *encryption.CMEKService {
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	s, err := encryption.NewCMEKService(ctx, gcs)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
