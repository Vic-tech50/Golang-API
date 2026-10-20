package main

import (
	"crypto/rand"     // Generates cryptographically secure random bytes.
	"encoding/base64" // Converts bytes into Base64 text.
	"fmt"             // Prints the generated key.
	"log"             // Reports errors.
)

// change mainly to main
func mainly() {
	// Create a 32-byte key for AES-256 encryption.
	key := make([]byte, 32)

	// Fill the key with secure random bytes.
	if _, err := rand.Read(key); err != nil {
		log.Fatal("Failed to generate key:", err)
	}

	// Encode the key so it can be stored in a .env file.
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Print the key for you to copy.
	fmt.Println("Your encryption key:")
	fmt.Println(encodedKey)
}
