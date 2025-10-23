package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Password Security (CONFIDENTIALITY)
// ============================================================================

const (
	// bcrypt cost (higher = slower but more secure; 12 is standard)
	bcryptCost = 12
)

// HashPassword securely hashes a password using bcrypt
// CONFIDENTIALITY: Never store plain passwords
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		log.Printf("[SECURITY] Error hashing password: %v", err)
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword compares a plain password with its hash
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
