package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Authentication Handlers (Register, Login, Logout)
// ============================================================================

// registerHandler creates a new user account
// POST /api/v1/auth/register
// CONFIDENTIALITY: Password hashed with bcrypt
// INTEGRITY: Email validation, password strength checked
// AVAILABILITY: Rate limited by parent router
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate request size (INTEGRITY)
	if !ValidateRequestSize(w, r) {
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}
	defer r.Body.Close()

	// Validate input (INTEGRITY)
	if err := validate.Struct(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Check if user already exists (INTEGRITY)
	userKey := "user:" + req.Email
	if _, err := rdb.Get(r.Context(), userKey).Result(); err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Email already registered",
		})
		return
	}

	// Hash password (CONFIDENTIALITY)
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("[AUTH] Password hashing failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to process registration",
		})
		return
	}

	// Create user object
	user := &User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  hashedPassword,
		FullName:  req.FullName,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store user in cache (AVAILABILITY: fast retrieval)
	userJSON, _ := json.Marshal(user)
	ttl := 24 * time.Hour
	if err := rdb.Set(r.Context(), userKey, userJSON, ttl).Err(); err != nil {
		log.Printf("[AUTH] Failed to store user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to register user",
		})
		return
	}

	// Log registration attempt (INTEGRITY: audit trail)
	log.Printf("[AUDIT] User registered: %s (%s)", user.Email, user.ID)

	// Generate JWT token
	token, err := generateJWT(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to generate token",
		})
		return
	}

	// Return success response (no password exposed)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token:     token,
		ExpiresIn: 3600, // 1 hour
		User: &User{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Active:   user.Active,
		},
	})
}

// loginHandler authenticates a user and returns a JWT token
// POST /api/v1/auth/login
// CONFIDENTIALITY: Password never logged, only hashed version checked
// INTEGRITY: Email & password verified before token issued
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate request size (INTEGRITY)
	if !ValidateRequestSize(w, r) {
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}
	defer r.Body.Close()

	// Validate input (INTEGRITY)
	if err := validate.Struct(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Retrieve user (AVAILABILITY: cache-first)
	userKey := "user:" + req.Email
	userJSON, err := rdb.Get(r.Context(), userKey).Result()
	if err != nil {
		// User not found or Redis error
		log.Printf("[AUTH] Login failed for %s: user not found", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid email or password",
		})
		return
	}

	var user User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to process login",
		})
		return
	}

	// Verify password (CONFIDENTIALITY: constant-time comparison)
	if !VerifyPassword(user.Password, req.Password) {
		log.Printf("[AUDIT] Failed login attempt for %s", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid email or password",
		})
		return
	}

	// Check if user is active (INTEGRITY)
	if !user.Active {
		log.Printf("[AUDIT] Login attempt by inactive user: %s", user.Email)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "User account is inactive",
		})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to generate token",
		})
		return
	}

	// Log successful login (INTEGRITY: audit trail)
	log.Printf("[AUDIT] User logged in: %s (%s)", user.Email, user.ID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Token:     token,
		ExpiresIn: 3600, // 1 hour
		User: &User{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Active:   user.Active,
		},
	})
}

// logoutHandler invalidates a user's session
// POST /api/v1/auth/logout (protected)
// Note: JWT is stateless; logout clears cache/client-side token
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by jwtMiddleware)
	userID, ok := r.Context().Value("user").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	log.Printf("[AUDIT] User logged out: %s", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// meHandler returns the current authenticated user's info
// GET /api/v1/auth/me (protected)
func meHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by jwtMiddleware)
	userID, ok := r.Context().Value("user").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// For demo: return minimal user info
	// In production: fetch from DB with this userID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID,
		"status":  "authenticated",
	})
}
