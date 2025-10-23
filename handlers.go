package main

import (
	"encoding/json"
	"net/http"
)

// legacyLoginHandler is the old simple demo login (kept for backward compatibility)
// POST /login - legacy endpoint
// Note: Use /api/v1/auth/login instead in new code
func legacyLoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		User string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.User == "" {
		http.Error(w, "invalid body (expect {\"user\":\"...\"})", http.StatusBadRequest)
		return
	}
	token, err := generateJWT(payload.User)
	if err != nil {
		http.Error(w, "failed create token", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// createUserHandler protected: memvalidasi input dan return success
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var in UserInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if !validateInput(w, &in) {
		return
	}
	// contoh: dapatkan user dari context (subject dari JWT)
	sub, _ := r.Context().Value("user").(string)
	_ = sub // di production, gunakan subject untuk otorisasi/owner check

	// simpan ke DB atau cache — disini cuma return
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "user created", "email": in.Email})
}
