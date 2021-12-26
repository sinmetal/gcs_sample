package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// GenerateEncryptionKey generates a 256 bit (32 byte) AES encryption key and
// prints the base64 representation.
func GenerateEncryptionKeyToWrite(w io.Writer) error {
	// This is included for demonstration purposes. You should generate your own
	// key. Please remember that encryption keys should be handled with a
	// comprehensive security policy.
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("rand.Read: %v", err)
	}
	encryptionKey := base64.StdEncoding.EncodeToString(key)
	fmt.Fprintf(w, "Generated base64-encoded encryption key: %v\n", encryptionKey)
	return nil
}

// GenerateEncryptionKey generates a 256 bit (32 byte) AES encryption key and
// prints the base64 representation.
func GenerateEncryptionKey() ([]byte, error) {
	// This is included for demonstration purposes. You should generate your own
	// key. Please remember that encryption keys should be handled with a
	// comprehensive security policy.
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("rand.Read: %v", err)
	}
	return key, nil
}
