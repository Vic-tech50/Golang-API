
package config

import (
	"crypto/aes"      // Provides AES encryption.
	"crypto/cipher"   // Provides authenticated encryption.
	"crypto/rand"     // Generates secure random bytes.
	"encoding/base64" // Converts encrypted bytes to text.
	"fmt"             // Creates descriptive errors.
	"io"              // Reads random bytes.
	"os"              // Reads the encryption key.
	"strings"         // Removes surrounding whitespace.
)

// getKey loads and validates the application's encryption key.
func getKey() ([]byte, error) {
	// Read the Base64-encoded key from the environment.
	encodedKey := strings.TrimSpace(os.Getenv("ENCRYPTION_KEY"))

	// Reject a missing key.
	if encodedKey == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is not configured")
	}

	// Decode the key from Base64.
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	// AES-256 requires exactly 32 bytes.
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	return key, nil
}

// Encrypt converts any string value into encrypted text.
func Encrypt(value string) (string, error) {
	// Load the secret key.
	key, err := getKey()
	if err != nil {
		return "", err
	}

	// Create an AES cipher using the secret key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM for authenticated encryption.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a unique nonce for this encryption.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the value and attach the authentication tag.
	// The nonce is prepended so Decrypt can recover it.
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)

	// Encode the encrypted bytes into printable text.
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}


func Decrypt(encryptedValue string) (string, error) {
	// Load the secret key used for encryption.
	key, err := getKey()
	if err != nil {
		return "", err
	}

	// Decode the Base64-encoded value.
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		// The value may be old plain-text data.
		// Return it unchanged instead of failing.
		return encryptedValue, nil
	}

	// Recreate the AES cipher.
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM for authenticated decryption.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Check whether the data is long enough to contain a nonce and tag.
	if len(ciphertext) < gcm.NonceSize()+gcm.Overhead() {
		// It is not a valid encrypted value in this format.
		return encryptedValue, nil
	}

	// Separate the nonce from the encrypted content.
	nonce := ciphertext[:gcm.NonceSize()]
	encryptedData := ciphertext[gcm.NonceSize():]

	// Attempt to decrypt and verify the data.
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		// Decryption failed; preserve the stored value.
		return encryptedValue, nil
	}

	// Return the successfully decrypted original value.
	return string(plaintext), nil
}


// Decrypt restores the original string from encrypted text.
// func Decrypt(encryptedValue string) (string, error) {
// 	// Load the same secret key used during encryption.
// 	key, err := getKey()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Decode the Base64-encoded ciphertext.
// 	ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
// 	if err != nil {
// 		return "", fmt.Errorf("invalid encrypted text: %w", err)
// 	}

// 	// Recreate the AES cipher.
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Create GCM for authenticated decryption.
// 	gcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Check that the ciphertext contains a complete nonce.
// 	if len(ciphertext) < gcm.NonceSize() {
// 		return "", fmt.Errorf("encrypted value is too short")
// 	}

// 	// Separate the nonce from the encrypted content.
// 	nonce := ciphertext[:gcm.NonceSize()]
// 	encryptedData := ciphertext[gcm.NonceSize():]

// 	// Decrypt the value and verify its authenticity.
// 	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
// 	if err != nil {
// 		return "", fmt.Errorf("decryption failed: %w", err)
// 	}

// 	// Return the original string.
// 	return string(plaintext), nil
// }
