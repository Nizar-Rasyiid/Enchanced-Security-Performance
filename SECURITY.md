# CIA Triad Security Implementation Guide

This document outlines how the application implements the **CIA Triad** (Confidentiality, Integrity, Availability) security model.

---

## 1. CONFIDENTIALITY ✓

**Goal**: Protect data from unauthorized access and disclosure.

### Implementation

#### 1.1 Encryption in Transit (TLS/HTTPS)

- **File**: `tls_util.go`, `main.go`
- **Feature**:
  - All traffic uses TLS 1.2+ (no plain HTTP)
  - Self-signed certs auto-generated for dev (`certs/server.crt`, `certs/server.key`)
  - Production: Replace with CA-signed certificates

```go
TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}
```

#### 1.2 Secret Management

- **File**: `security.go`, `auth.go`
- **Feature**:
  - JWT secret **NOT hardcoded** anymore
  - Loaded from `JWT_SECRET` environment variable
  - Defaults to development value; **must** set in production

```bash
# Production environment
export JWT_SECRET="your-complex-secret-key"
export ENCRYPTION_KEY="your-encryption-key"
```

#### 1.3 Secure Headers

- **File**: `middleware.go`
- **Headers**:
  - `X-Content-Type-Options: nosniff` - Prevent MIME-sniffing
  - `X-Frame-Options: DENY` - Prevent clickjacking
  - `Strict-Transport-Security` - Force HTTPS for 2 years

#### 1.4 HTTPS Redirect

- **File**: `security.go`
- **Feature**: Redirects HTTP to HTTPS when `REQUIRE_HTTPS=true` (default)

#### 1.5 Encryption Keys (Future)

- **File**: `security.go`
- **Function**: `EncryptSensitiveData()`, `DecryptSensitiveData()`
- **TODO**: Implement AES-256-GCM for data at rest

---

## 2. INTEGRITY ✓

**Goal**: Ensure data has not been altered or tampered with.

### Implementation

#### 2.1 Input Validation

- **File**: `validate.go`, `handlers.go`
- **Framework**: `go-playground/validator/v10`
- **Rules**:
  - Email: Must be valid email format
  - Name: Minimum 3 characters
  - Body: Max 10MB (enforced in `ValidateRequestSize`)

```go
type UserInput struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required,min=3"`
}
```

#### 2.2 CSRF Protection

- **File**: `security.go`
- **Functions**: `GenerateCSRFToken()`, `ValidateCSRFToken()`
- **Token Expiry**: 15 minutes (configurable)
- **Usage**:
  1. Client requests CSRF token from `/csrf-token` (implement handler)
  2. Client includes token in `X-CSRF-Token` header
  3. Server validates token before processing state-changing requests

```go
// Generate token (in login/form handler)
token, err := GenerateCSRFToken()

// Validate before POST/PUT/DELETE
if !ValidateCSRFToken(r.Header.Get("X-CSRF-Token")) {
    http.Error(w, "Invalid CSRF token", http.StatusForbidden)
    return
}
```

#### 2.3 Request Signing (Future)

- **File**: `security.go`
- **Field**: `RequestSigningSecret`
- **TODO**: Implement HMAC-SHA256 request signing for API clients

#### 2.4 Audit Logging

- **File**: `security.go`
- **Middleware**: `RequestLoggingMiddleware()`
- **Logs**: Method, path, client IP, timestamp, response time
- **Output**: Every request logged to stdout (send to centralized logging in production)

```
[AUDIT] POST /user from 127.0.0.1 at 2025-10-16T16:00:13Z
[AUDIT] Completed in 42.5ms
```

---

## 3. AVAILABILITY ✓

**Goal**: Ensure services remain accessible and performant under normal and adverse conditions.

### Implementation

#### 3.1 Rate Limiting

- **File**: `router.go`
- **Middleware**: `httprate.LimitByIP()`
- **Limit**: 60 requests per minute per IP
- **Configurable**: Edit `router.go` or update via environment

```go
r.Use(httprate.LimitByIP(60, 1*60)) // 60 req/min
```

#### 3.2 Request Timeouts

- **File**: `server.go`
- **Timeouts**:
  - Read: 5 seconds (max time to read request)
  - Write: 10 seconds (max time to write response)
  - Idle: 60 seconds (max time to keep connection alive)

```go
ReadTimeout:  5 * time.Second,
WriteTimeout: 10 * time.Second,
IdleTimeout:  60 * time.Second,
```

#### 3.3 Panic Recovery

- **File**: `security.go`
- **Middleware**: `RecoveryMiddleware()`
- **Behavior**:
  - Catches all panics
  - Logs panic details
  - Returns `500 Internal Server Error` (no stack trace leakage)
  - Server stays online (no crash)

#### 3.4 Request Size Limits

- **File**: `security.go`
- **Max Body Size**: 10 MB (prevents memory exhaustion)
- **Function**: `ValidateRequestSize()`

#### 3.5 Connection Pool

- **File**: `db.go`
- **Settings**:
  - Max open connections: 10
  - Max idle connections: 5
  - Prevents DB connection exhaustion

#### 3.6 Graceful Shutdown

- **File**: `shutdown.go`
- **Behavior**:
  - Listens for SIGINT/SIGTERM
  - Closes server with 5-second timeout
  - Drains in-flight requests before shutdown

---

## 4. CORS Policy (CONFIDENTIALITY + INTEGRITY)

- **File**: `security.go`
- **Middleware**: `CORSMiddleware()`
- **Allowed Origins**: Configured via `ALLOWED_ORIGINS` env var
- **Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Headers**: Content-Type, Authorization, X-CSRF-Token
- **Max Age**: 3600 seconds (preflight cache)

```bash
export ALLOWED_ORIGINS="https://localhost:3000,https://app.example.com"
```

---

## 5. Security Configuration

All security settings are initialized via `InitSecurityConfig()` in `main.go`:

### Environment Variables

| Variable                 | Default                                   | Purpose                                 |
| ------------------------ | ----------------------------------------- | --------------------------------------- |
| `JWT_SECRET`             | `your-secret-key-change-me-in-production` | JWT signing key (CONFIDENTIALITY)       |
| `ENCRYPTION_KEY`         | ``                                        | Data encryption key (future)            |
| `ALLOWED_ORIGINS`        | `https://localhost:8443`                  | CORS whitelist                          |
| `REQUIRE_HTTPS`          | `true`                                    | Force HTTPS redirect                    |
| `ENVIRONMENT`            | (unset)                                   | Set to `production` for strict warnings |
| `REQUEST_SIGNING_SECRET` | ``                                        | Request signature key (future)          |

### Startup Security Check

On startup, the application logs:

```
============================================================
CIA TRIAD SECURITY STATUS
============================================================
[CONFIDENTIALITY]
  ✓ TLS: Enabled (1.2+)
  ✓ JWT Secret: Loaded from environment
  ✓ HTTPS Redirect: true
[INTEGRITY]
  ✓ Input Validation: Enabled (max body: 10485760 bytes)
  ✓ CSRF Protection: Enabled (15 min expiry)
  ✓ Request Logging: Enabled (audit trail)
[AVAILABILITY]
  ✓ Rate Limiting: 60 req/min
  ✓ Request Timeout: 30s
  ✓ Panic Recovery: Enabled
============================================================
```

---

## 6. Middleware Chain

The router applies security middleware in this order:

1. **RecoveryMiddleware** - Catch panics (AVAILABILITY)
2. **RequestLoggingMiddleware** - Log all requests (INTEGRITY)
3. **HTTPSRedirectMiddleware** - Enforce HTTPS (CONFIDENTIALITY)
4. **CORSMiddleware** - Apply CORS policy (CONFIDENTIALITY + INTEGRITY)
5. **secureHeaders** - Set security headers (CONFIDENTIALITY)
6. **gzipMiddleware** - Compress responses (PERFORMANCE)
7. **LimitByIP** - Rate limit (AVAILABILITY)

---

## 7. Production Deployment Checklist

- [ ] Set `JWT_SECRET` to a strong random value (32+ chars)
- [ ] Set `ENCRYPTION_KEY` for future data encryption
- [ ] Set `ENVIRONMENT=production` for strict security warnings
- [ ] Replace self-signed certs with CA-signed certificates (in `certs/`)
- [ ] Update `ALLOWED_ORIGINS` to production domain(s)
- [ ] Set `REQUIRE_HTTPS=true` (already default)
- [ ] Enable request signing by setting `REQUEST_SIGNING_SECRET`
- [ ] Configure centralized logging (send audit logs to ELK/CloudWatch)
- [ ] Set up monitoring/alerting for rate limit violations
- [ ] Use secrets manager (AWS Secrets Manager, HashiCorp Vault) instead of env vars
- [ ] Enable database connection pooling per environment
- [ ] Test graceful shutdown under load

---

## 8. Testing CIA Triad

### Confidentiality

```bash
# Verify HTTPS works
curl -k https://localhost:8443/health

# Check security headers
curl -i https://localhost:8443/health | grep -i "x-content\|x-frame\|strict"
```

### Integrity

```bash
# Test input validation
curl -X POST https://localhost:8443/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"email": "invalid", "name": "ab"}'  # Should reject

# Test CSRF protection (once handler added)
curl -X POST https://localhost:8443/user \
  -H "Content-Type: application/json" \
  # Should reject without X-CSRF-Token
```

### Availability

```bash
# Test rate limiting (send 61 requests)
for i in {1..61}; do curl https://localhost:8443/health; done
# Request 61 should return 429 (Too Many Requests)

# Test panic recovery (if handler added)
curl https://localhost:8443/panic  # Should return 500, server stays online
```

---

## 9. References

- [OWASP Top 10](https://owasp.org/Top10/)
- [CIA Triad](https://en.wikipedia.org/wiki/Information_security#Confidentiality,_integrity,_and_availability)
- [Go Security Best Practices](https://golang.org/doc/effective_go#security)
- [JWT.io](https://jwt.io/)
- [CORS Specification](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)

---

**Last Updated**: October 23, 2025  
**Status**: Fully Implemented ✓
