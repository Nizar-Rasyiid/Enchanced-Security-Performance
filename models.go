package main

import "time"

// ============================================================================
// User Models (Authentication)
// ============================================================================

// User represents a registered user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	FullName  string    `json:"full_name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterRequest is the payload for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=3"`
}

// LoginRequest is the payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned after successful login/register
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // seconds
	User      *User  `json:"user"`
}

// ============================================================================
// Health Data Models
// ============================================================================

// HealthRecord represents a single health measurement record
type HealthRecord struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Type       string    `json:"type"` // "blood_pressure", "heart_rate", "weight", "temperature", "glucose"
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"` // "mmHg", "bpm", "kg", "°C", "mg/dL"
	Notes      string    `json:"notes"`
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// HealthRecordRequest is the payload for creating/updating health records
type HealthRecordRequest struct {
	Type       string  `json:"type" validate:"required,oneof=blood_pressure heart_rate weight temperature glucose"`
	Value      float64 `json:"value" validate:"required,min=0"`
	Unit       string  `json:"unit" validate:"required"`
	Notes      string  `json:"notes" validate:"max=500"`
	RecordedAt string  `json:"recorded_at"` // ISO 8601 format
}

// HealthStats represents aggregated health statistics
type HealthStats struct {
	UserID     string    `json:"user_id"`
	Type       string    `json:"type"`
	Average    float64   `json:"average"`
	Min        float64   `json:"min"`
	Max        float64   `json:"max"`
	Count      int       `json:"count"`
	LastRecord time.Time `json:"last_record"`
}

// ============================================================================
// Validation Structs
// ============================================================================

// HealthRecordInput for generic validation
type HealthRecordInput struct {
	Type  string  `validate:"required,oneof=blood_pressure heart_rate weight temperature glucose"`
	Value float64 `validate:"required,min=0,max=500"`
	Unit  string  `validate:"required"`
}
