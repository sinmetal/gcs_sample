package encryption

import (
	"context"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/sinmetal/gcs_sample/internal/trace"
)

type CMEKService struct {
	gcs *storage.Client
}

func NewCMEKService(ctx context.Context, gcs *storage.Client) (*CMEKService, error) {
	return &CMEKService{
		gcs: gcs,
	}, nil
}

// Upload is Cloud Storageにfileをアップロードする
// CMEKとしてBucket Default Keyを指定しているので、コード上はただアップロードしてるだけ
func (s *CMEKService) Upload(ctx context.Context, bucketName string, objectName string, file []byte) (size int, err error) {
	ctx = trace.StartSpan(ctx, "encryption/cmek/upload")
	defer trace.EndSpan(ctx, err)

	// bucket default keyを指定してるので、普通にUploadしている
	// https://cloud.google.com/storage/docs/encryption/using-customer-managed-keys?hl=en#add-default-key
	obj := s.gcs.Bucket(bucketName).Object(objectName)
	w := obj.NewWriter(ctx)

	size, err = w.Write(file)
	if err != nil {
		return 0, fmt.Errorf("failed gcs.write: %w", err)
	}

	if err := w.Close(); err != nil {
		return size, fmt.Errorf("file writer close error: %w", err)
	}

	return size, nil
}

// Download is Cloud Storageからobjectをダウンロードする
// CMEKとしてBucket Default Keyを指定しているので、コード上はただダウンロードしてるだけ
func (s *CMEKService) Download(ctx context.Context, keyName string, bucketName string, objectName string) (data []byte, err error) {
	ctx = trace.StartSpan(ctx, "encryption/cmek/download")
	defer trace.EndSpan(ctx, err)

	obj := s.gcs.Bucket(bucketName).Object(objectName)
	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed object.NewReader: %w", err)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			// noop
		}
	}()

	data, err = ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed object.Read: %w", err)
	}
	return data, nil
}

// ReEncrypt is KeyをRotateした後に、新しいKeyでEncryptし直す時に利用する
// Bucket Default Keyとして設定しているKeyをRotationした後、実行することを想定しているので、実際やっていることはobjectを同じPathにCopyしているだけ
func (s *CMEKService) ReEncrypt(ctx context.Context, bucketName string, objectName string) (err error) {
	ctx = trace.StartSpan(ctx, "encryption/cmek/reEncrypt")
	defer trace.EndSpan(ctx, err)

	src := s.gcs.Bucket(bucketName).Object(objectName)

	// 同じObject PathにCopyする
	// Object Pathが同一でも実際には別のObjectになるので、Copyが成功すれば新しいObjectが返されるようになり、Copy中およびCopyが失敗した場合は元のObjectが返される状態が維持される
	copier := s.gcs.Bucket(bucketName).Object(objectName).CopierFrom(src)
	_, err = copier.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}
