package encryption

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"google.golang.org/api/cloudkms/v1"
)

// CSEKService is customer-supplied encryption keys Service
type CSEKService struct {
	gcs *storage.Client
	kms *cloudkms.Service
}

func NewCSEKService(ctx context.Context, gcs *storage.Client, kms *cloudkms.Service) (*CSEKService, error) {
	return &CSEKService{
		gcs: gcs,
		kms: kms,
	}, nil
}

// Encrypt is 指定したCloud KMSの鍵で暗号化する
// keyName format: "projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s
func (s *CSEKService) Encrypt(ctx context.Context, keyName string, plaintext string) (ciphertext string, cryptoKey string, err error) {
	response, err := s.kms.Projects.Locations.KeyRings.CryptoKeys.Encrypt(keyName, &cloudkms.EncryptRequest{
		Plaintext: plaintext,
	}).Do()
	if err != nil {
		return "", "", fmt.Errorf("encrypt: failed to encrypt. CryptoKey=%s : %w", keyName, err)
	}

	return response.Ciphertext, response.Name, nil
}

func (s *CSEKService) Decrypt(ctx context.Context, keyName string, ciphertext string) (plaintext string, err error) {
	response, err := s.kms.Projects.Locations.KeyRings.CryptoKeys.Decrypt(keyName, &cloudkms.DecryptRequest{
		Ciphertext: ciphertext,
	}).Do()
	if err != nil {
		return "", fmt.Errorf("decrypt: failed to decrypt. CryptoKey=%s : %w", keyName, err)
	}

	return response.Plaintext, nil
}

// Upload is Cloud Storageに指定されたファイルをアップロードする
// アップロードする時にcustomer-supplied encryption keyとしてencryptionKeyを利用する
// encryptionKeyはkeyNameで指定されたCloud KMS Keyを利用して暗号化し、Object.Metadata[wDEK]として保存する
//
// keyName format: "projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s
// encryptionKey: 256 bit (32 byte) AES encryption key
func (s *CSEKService) Upload(ctx context.Context, keyName string, bucketName string, objectName string, encryptionKey []byte, file []byte) (int, error) {
	obj := s.gcs.Bucket(bucketName).Object(objectName).Key(encryptionKey)
	w := obj.NewWriter(ctx)

	ekt := base64.StdEncoding.EncodeToString(encryptionKey)
	chiphertext, cryptKey, err := s.Encrypt(ctx, keyName, ekt)
	if err != nil {
		return 0, fmt.Errorf("failed encrypt: %w", err)
	}

	metadata := map[string]string{}
	metadata["wDEK"] = chiphertext
	metadata["cryptKey"] = cryptKey
	w.Metadata = metadata
	size, err := w.Write(file)
	if err != nil {
		return 0, fmt.Errorf("failed gcs.write: %w", err)
	}

	if err := w.Close(); err != nil {
		return size, fmt.Errorf("file writer close error: %w", err)
	}

	return size, nil
}

// Download is Cloud Storageから指定されたファイルをダウンロードする
// ダウンロードする時にcustomer-supplied encryption keyとして、Object.Metadata[wDEK]から取得した値をkeyNameで指定されたCloud KMS Keyで復号化して利用する
//
// keyName format: "projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s
func (s *CSEKService) Download(ctx context.Context, keyName string, bucketName string, objectName string) ([]byte, error) {
	obj := s.gcs.Bucket(bucketName).Object(objectName)
	attrs, err := obj.Attrs(ctx)
	encryptedSecretKey := attrs.Metadata["wDEK"]
	if len(encryptedSecretKey) < 1 {
		return nil, fmt.Errorf("not found encryptedSecretKey from object.Metadata[wDEK]")
	}

	plainttext, err := s.Decrypt(ctx, keyName, encryptedSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed decrpyt encryptedSecretKey: %w", err)
	}
	secretKey, err := base64.StdEncoding.DecodeString(plainttext)
	if err != nil {
		return nil, fmt.Errorf("failed base64.Decode encryptedSecretKey: %w", err)
	}

	rc, err := obj.Key(secretKey).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed object.NewReader: %w", err)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			// noop
		}
	}()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed object.Read: %w", err)
	}
	return data, nil
}