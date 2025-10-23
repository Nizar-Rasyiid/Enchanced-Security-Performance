package main

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

func setupRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()

	// security middleware (CIA triad)
	r.Use(RecoveryMiddleware)       // AVAILABILITY: panic recovery
	r.Use(RequestLoggingMiddleware) // INTEGRITY: audit trail
	r.Use(HTTPSRedirectMiddleware)  // CONFIDENTIALITY: enforce HTTPS
	r.Use(CORSMiddleware)           // CONFIDENTIALITY + INTEGRITY: CORS policy

	// security & performance middleware
	r.Use(secureHeaders)
	r.Use(gzipMiddleware)
	// rate limit: 60 req per minute per client (tweak sesuai kebutuhan)
	r.Use(httprate.LimitByIP(60, 1*60))

	// public endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public authentication endpoints
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", registerHandler)
			r.Post("/login", loginHandler)
		})

		// Protected endpoints (require JWT)
		r.Group(func(rg chi.Router) {
			rg.Use(jwtMiddleware)

			// User auth endpoints
			rg.Route("/auth", func(r chi.Router) {
				r.Post("/logout", logoutHandler)
				r.Get("/me", meHandler)
			})

			// Health data endpoints (CRUD)
			rg.Route("/health", func(r chi.Router) {
				r.Post("/", createHealthRecordHandler)
				r.Get("/", getHealthRecordsHandler)
				r.Get("/stats", getHealthStatsHandler)
				r.Delete("/", deleteHealthRecordHandler)
			})
		})
	})

	// Legacy endpoints (for backward compatibility)
	r.Post("/login", legacyLoginHandler)
	r.Group(func(rg chi.Router) {
		rg.Use(jwtMiddleware)
		rg.Post("/user", createUserHandler)
	})

	// optional: health/metrics etc.
	_ = db // agar param used jika diperlukan
	return r
}
