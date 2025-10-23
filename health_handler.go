package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Health Data Handlers (CRUD operations)
// ============================================================================

// createHealthRecordHandler creates a new health record for the authenticated user
// POST /api/v1/health (protected)
// CONFIDENTIALITY: User data isolated by user_id
// INTEGRITY: Input validated, timestamps enforced
// AVAILABILITY: Cached for fast reads
func createHealthRecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate request size (INTEGRITY)
	if !ValidateRequestSize(w, r) {
		return
	}

	// Get user ID from context (set by jwtMiddleware)
	userID, ok := r.Context().Value("user").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var req HealthRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}
	defer r.Body.Close()

	// Validate input (INTEGRITY)
	input := HealthRecordInput{
		Type:  req.Type,
		Value: req.Value,
		Unit:  req.Unit,
	}
	if err := validate.Struct(input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Parse recorded time
	recordedAt := time.Now()
	if req.RecordedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.RecordedAt); err == nil {
			recordedAt = t
		}
	}

	// Create health record
	record := &HealthRecord{
		ID:         uuid.New().String(),
		UserID:     userID,
		Type:       req.Type,
		Value:      req.Value,
		Unit:       req.Unit,
		Notes:      req.Notes,
		RecordedAt: recordedAt,
		CreatedAt:  time.Now(),
	}

	// Store in cache (AVAILABILITY: fast retrieval)
	recordKey := fmt.Sprintf("health:%s:%s", userID, record.ID)
	recordJSON, _ := json.Marshal(record)
	ttl := 30 * 24 * time.Hour // 30 days
	if err := rdb.Set(r.Context(), recordKey, recordJSON, ttl).Err(); err != nil {
		log.Printf("[HEALTH] Failed to store record: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create record"})
		return
	}

	// Add to user's health record list (for indexing)
	listKey := fmt.Sprintf("health:%s:list", userID)
	rdb.LPush(r.Context(), listKey, record.ID)
	rdb.Expire(r.Context(), listKey, ttl)

	// Invalidate stats cache (AVAILABILITY: invalidate on write)
	statsKey := fmt.Sprintf("health:%s:stats:%s", userID, req.Type)
	rdb.Del(r.Context(), statsKey)

	log.Printf("[AUDIT] Health record created: %s for user: %s", record.ID, userID)

	// Return created record
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// getHealthRecordsHandler retrieves health records for the authenticated user
// GET /api/v1/health (protected)
// AVAILABILITY: Cached response, pagination support
func getHealthRecordsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("user").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Get pagination params
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit > 100 {
			limit = 100 // Max 100 to prevent DoS
		}
	}

	// Retrieve from cache (AVAILABILITY: fast reads with caching)
	listKey := fmt.Sprintf("health:%s:list", userID)
	recordIDs, err := rdb.LRange(r.Context(), listKey, 0, int64(limit-1)).Result()
	if err != nil {
		// If list not found, return empty
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]HealthRecord{})
		return
	}

	var records []*HealthRecord
	for _, id := range recordIDs {
		recordKey := fmt.Sprintf("health:%s:%s", userID, id)
		recordJSON, err := rdb.Get(r.Context(), recordKey).Result()
		if err == nil {
			var record HealthRecord
			if err := json.Unmarshal([]byte(recordJSON), &record); err == nil {
				records = append(records, &record)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", fmt.Sprintf("%d", len(records)))
	w.WriteHeader(http.StatusOK)
	if records == nil {
		records = []*HealthRecord{}
	}
	json.NewEncoder(w).Encode(records)
}

// getHealthStatsHandler returns aggregated stats for a specific health metric type
// GET /api/v1/health/stats?type=heart_rate (protected)
// AVAILABILITY: Cached aggregation results
func getHealthStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("user").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Get type parameter
	recordType := r.URL.Query().Get("type")
	if recordType == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing 'type' parameter"})
		return
	}

	// Check cache first (AVAILABILITY: fast aggregation)
	statsKey := fmt.Sprintf("health:%s:stats:%s", userID, recordType)
	statsJSON, err := rdb.Get(r.Context(), statsKey).Result()
	if err == nil {
		// Cache hit
		var stats HealthStats
		if err := json.Unmarshal([]byte(statsJSON), &stats); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(stats)
			return
		}
	}

	// Cache miss: compute stats from all records
	listKey := fmt.Sprintf("health:%s:list", userID)
	recordIDs, _ := rdb.LRange(r.Context(), listKey, 0, -1).Result()

	var values []float64
	var lastRecord time.Time

	for _, id := range recordIDs {
		recordKey := fmt.Sprintf("health:%s:%s", userID, id)
		recordJSON, err := rdb.Get(r.Context(), recordKey).Result()
		if err == nil {
			var record HealthRecord
			if err := json.Unmarshal([]byte(recordJSON), &record); err == nil {
				if record.Type == recordType {
					values = append(values, record.Value)
					if record.RecordedAt.After(lastRecord) {
						lastRecord = record.RecordedAt
					}
				}
			}
		}
	}

	// Calculate aggregates
	var stats HealthStats
	if len(values) > 0 {
		stats = HealthStats{
			UserID:     userID,
			Type:       recordType,
			Count:      len(values),
			Average:    calculateAverage(values),
			Min:        calculateMin(values),
			Max:        calculateMax(values),
			LastRecord: lastRecord,
		}

		// Cache stats for 1 hour (AVAILABILITY)
		statsJSON, _ := json.Marshal(stats)
		rdb.Set(r.Context(), statsKey, statsJSON, 1*time.Hour)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// deleteHealthRecordHandler deletes a specific health record
// DELETE /api/v1/health/:id (protected)
func deleteHealthRecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("user").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Get record ID from URL
	recordID := r.URL.Query().Get("id")
	if recordID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing 'id' parameter"})
		return
	}

	// Delete from cache
	recordKey := fmt.Sprintf("health:%s:%s", userID, recordID)
	if err := rdb.Del(r.Context(), recordKey).Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete record"})
		return
	}

	// Remove from list index
	listKey := fmt.Sprintf("health:%s:list", userID)
	rdb.LRem(r.Context(), listKey, 1, recordID)

	log.Printf("[AUDIT] Health record deleted: %s for user %s", recordID, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record deleted successfully"})
}

// ============================================================================
// Helper functions for statistics
// ============================================================================

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}
