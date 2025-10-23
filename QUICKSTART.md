# Quick Reference Guide

## Project Setup

```powershell
# Navigate to project
cd "d:\Tugas Kuliah\Skripsi\golang-enchanced"

# Download dependencies
go mod download

# Build application
go build -o bin/server.exe

# Run application
go run .

# Run tests
go test ./... -v

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

---

## API Quick Commands

### 1. Start Server

```powershell
# Development (uses default JWT secret)
go run .

# Or with explicit secret
$env:JWT_SECRET = "my-secret-key"
go run .

# Build and run
go build -o bin/server.exe
.\bin\server.exe
```

### 2. Register User

```powershell
$registerResponse = curl -k -X POST "https://localhost:8443/api/v1/auth/register" `
  -H "Content-Type: application/json" `
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!",
    "full_name": "John Doe"
  }' | ConvertFrom-Json

$token = $registerResponse.token
echo "Token: $token"
```

### 3. Login User

```powershell
$loginResponse = curl -k -X POST "https://localhost:8443/api/v1/auth/login" `
  -H "Content-Type: application/json" `
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }' | ConvertFrom-Json

$token = $loginResponse.token
```

### 4. Create Health Record

```powershell
curl -k -X POST "https://localhost:8443/api/v1/health" `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{
    "type": "heart_rate",
    "value": 72.5,
    "unit": "bpm",
    "notes": "Morning measurement"
  }'
```

### 5. Get Health Records

```powershell
curl -k -X GET "https://localhost:8443/api/v1/health" `
  -H "Authorization: Bearer $token"
```

### 6. Get Health Statistics

```powershell
curl -k -X GET "https://localhost:8443/api/v1/health/stats?type=heart_rate" `
  -H "Authorization: Bearer $token" `
  -i  # Show headers including X-Cache
```

### 7. Delete Health Record

```powershell
# First, get a record ID from GET /health, then:
curl -k -X DELETE "https://localhost:8443/api/v1/health?id=<record-id>" `
  -H "Authorization: Bearer $token"
```

### 8. Logout

```powershell
curl -k -X POST "https://localhost:8443/api/v1/auth/logout" `
  -H "Authorization: Bearer $token"
```

### 9. Get Current User

```powershell
curl -k -X GET "https://localhost:8443/api/v1/auth/me" `
  -H "Authorization: Bearer $token"
```

---

## Full Testing Flow

```powershell
# 1. Register new user and save token
$registerResponse = curl -k -X POST "https://localhost:8443/api/v1/auth/register" `
  -H "Content-Type: application/json" `
  -d @'{
    "email": "testuser@example.com",
    "password": "Password123!",
    "full_name": "Test User"
  }' | ConvertFrom-Json

$token = $registerResponse.token
Write-Host "Token: $token" -ForegroundColor Green

# 2. Create multiple health records
$healthTypes = @("heart_rate", "blood_pressure", "temperature")
$values = @(72.5, 120, 37.2)

for ($i = 0; $i -lt 3; $i++) {
    curl -k -X POST "https://localhost:8443/api/v1/health" `
      -H "Authorization: Bearer $token" `
      -H "Content-Type: application/json" `
      -d @{
        "type" = $healthTypes[$i]
        "value" = $values[$i]
        "unit" = if ($i -eq 0) { "bpm" } elseif ($i -eq 1) { "mmHg" } else { "°C" }
      } | ConvertFrom-Json | Write-Host
}

# 3. Get all records
Write-Host "`n=== All Health Records ===" -ForegroundColor Cyan
curl -k -X GET "https://localhost:8443/api/v1/health" `
  -H "Authorization: Bearer $token" | ConvertFrom-Json | ConvertTo-Json

# 4. Get statistics
Write-Host "`n=== Heart Rate Stats ===" -ForegroundColor Cyan
curl -k -X GET "https://localhost:8443/api/v1/health/stats?type=heart_rate" `
  -H "Authorization: Bearer $token" | ConvertFrom-Json | ConvertTo-Json

# 5. Logout
curl -k -X POST "https://localhost:8443/api/v1/auth/logout" `
  -H "Authorization: Bearer $token"

Write-Host "`n✅ Test completed!" -ForegroundColor Green
```

---

## Development Workflow

### 1. Make Changes

```powershell
# Edit a file, e.g., health_handler.go
code health_handler.go
```

### 2. Rebuild & Run

```powershell
go build ./...
if ($?) { go run . }
```

### 3. Test Changes

```powershell
# Use the commands above to test new functionality
curl -k -X GET "https://localhost:8443/api/v1/health" `
  -H "Authorization: Bearer $token"
```

### 4. Format & Lint

```powershell
go fmt ./...
go vet ./...
```

---

## Monitoring

### Check Server Status

```powershell
curl -k https://localhost:8443/health
```

### View Response Headers

```powershell
curl -k -i https://localhost:8443/api/v1/auth/me `
  -H "Authorization: Bearer $token"

# Look for:
# X-Cache: HIT/MISS (cache status)
# Content-Encoding: gzip (compression)
# X-Content-Type-Options: nosniff (security)
```

### Performance Testing

```powershell
# Measure response time
$stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
curl -k -X GET "https://localhost:8443/api/v1/health/stats?type=heart_rate" `
  -H "Authorization: Bearer $token" | Out-Null
$stopwatch.Stop()
Write-Host "Response time: $($stopwatch.ElapsedMilliseconds)ms"

# Should be ~2-3ms from cache
```

---

## Troubleshooting

### Build Fails

```powershell
# Clean and rebuild
go clean ./...
go mod tidy
go build ./...
```

### Redis Connection Error

```powershell
# Check Redis is running
redis-cli ping

# If not running:
# On Windows: Open WSL/Docker and run: redis-server
# Or install Redis-Windows binary
```

### TLS Certificate Errors

```powershell
# Certificates are auto-generated on first run in ./certs/
# If missing, the app will recreate them

# To manually regenerate:
Remove-Item -Path "certs\" -Recurse -Force
go run .  # Will regenerate certs
```

### Port Already in Use

```powershell
# Find process using port 8443
Get-Process | Where-Object {$_.ProcessName -like "*go*"} | Stop-Process -Force

# Or use different port (edit server.go and router.go)
```

---

## Environment Setup

### Windows PowerShell

```powershell
# Set environment variables (session only)
$env:JWT_SECRET = "my-secret"
$env:ENVIRONMENT = "production"
$env:ALLOWED_ORIGINS = "https://app.example.com"

# Or set permanently (requires admin):
# [Environment]::SetEnvironmentVariable("JWT_SECRET", "my-secret", "User")
```

### Windows Command Prompt

```batch
set JWT_SECRET=my-secret
set ENVIRONMENT=production
go run .
```

### WSL/Linux/macOS

```bash
export JWT_SECRET="my-secret"
export ENVIRONMENT="production"
export ALLOWED_ORIGINS="https://app.example.com"
go run .
```

---

## File Structure Explained

| File                | Purpose                                      |
| ------------------- | -------------------------------------------- |
| `main.go`           | Entry point, initializes security & services |
| `router.go`         | API route definitions                        |
| `auth_handler.go`   | Register/login/logout endpoints              |
| `health_handler.go` | Health data CRUD & statistics                |
| `auth.go`           | JWT middleware & token generation            |
| `security.go`       | CIA triad framework & middleware             |
| `models.go`         | Data structures (User, HealthRecord)         |
| `password.go`       | Bcrypt password hashing utilities            |
| `validate.go`       | Input validation setup                       |
| `cache.go`          | Redis client initialization                  |
| `middleware.go`     | Security headers, gzip, etc.                 |
| `SECURITY.md`       | Complete security documentation              |
| `API.md`            | REST API reference                           |
| `README.md`         | Project overview                             |

---

## Key Metrics

| Metric                       | Value             |
| ---------------------------- | ----------------- |
| **Build Time**               | < 2 seconds       |
| **Startup Time**             | < 1 second        |
| **Auth Latency**             | ~50ms (bcrypt)    |
| **Cache Hit Latency**        | ~2-3ms            |
| **Stats Latency (cached)**   | ~3ms              |
| **Stats Latency (computed)** | ~50ms             |
| **Rate Limit**               | 60 req/min per IP |
| **Max Request Size**         | 10 MB             |
| **Cache TTL (users)**        | 24 hours          |
| **Cache TTL (health)**       | 30 days           |
| **Cache TTL (stats)**        | 1 hour            |

---

## Production Checklist

Before deploying to production:

```powershell
# 1. Security
- [ ] Set JWT_SECRET to strong random key
- [ ] Set ENVIRONMENT=production
- [ ] Replace self-signed certs
- [ ] Enable HTTPS only

# 2. Performance
- [ ] Setup Redis cluster
- [ ] Enable CDN/cache headers
- [ ] Load test the API

# 3. Operations
- [ ] Setup logging & monitoring
- [ ] Setup backups
- [ ] Setup CI/CD pipeline
- [ ] Create runbooks

# 4. Validation
- [ ] Penetration test
- [ ] Security audit
- [ ] Performance audit
- [ ] Load testing
```

---

**Version**: 1.0  
**Last Updated**: October 23, 2025  
**Maintained by**: Your Team
