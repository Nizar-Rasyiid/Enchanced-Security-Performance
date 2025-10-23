# API Documentation - Auth & Health Data System

This API provides secure user authentication and health data management with **CIA Triad** security and **Enhanced Performance**.

**Base URL**: `https://localhost:8443/api/v1`

---

## Table of Contents

1. [Authentication Endpoints](#authentication-endpoints)
2. [Health Data Endpoints](#health-data-endpoints)
3. [CIA Triad Implementation](#cia-triad-implementation)
4. [Performance Enhancements](#performance-enhancements)
5. [Error Handling](#error-handling)

---

## Authentication Endpoints

### 1. Register User

Create a new user account with secure password hashing.

**Endpoint**: `POST /auth/register`  
**Access**: Public  
**Security**: CONFIDENTIALITY (bcrypt password hashing), INTEGRITY (input validation)

**Request Body**:

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe"
}
```

**Request Headers**:

```
Content-Type: application/json
```

**Validation Rules**:

- `email`: Must be valid email format (required)
- `password`: Minimum 8 characters (required)
- `full_name`: Minimum 3 characters (required)

**Response** (201 Created):

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "full_name": "John Doe",
    "active": true
  }
}
```

**Error Responses**:

- `400 Bad Request`: Invalid input
- `409 Conflict`: Email already registered

**Example**:

```bash
curl -X POST https://localhost:8443/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "full_name": "John Doe"
  }'
```

---

### 2. Login

Authenticate user and receive JWT token.

**Endpoint**: `POST /auth/login`  
**Access**: Public  
**Security**: CONFIDENTIALITY (password verification), INTEGRITY (audit logging)

**Request Body**:

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response** (200 OK):

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "full_name": "John Doe",
    "active": true
  }
}
```

**Error Responses**:

- `401 Unauthorized`: Invalid email/password
- `403 Forbidden`: User account inactive

**Example**:

```bash
curl -X POST https://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

---

### 3. Get Current User

Retrieve authenticated user information.

**Endpoint**: `GET /auth/me`  
**Access**: Protected (requires valid JWT)  
**Security**: CONFIDENTIALITY (user isolation)

**Request Headers**:

```
Authorization: Bearer <token>
```

**Response** (200 OK):

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "authenticated"
}
```

**Example**:

```bash
curl -X GET https://localhost:8443/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

### 4. Logout

Invalidate user session (client-side token deletion).

**Endpoint**: `POST /auth/logout`  
**Access**: Protected (requires valid JWT)  
**Security**: INTEGRITY (audit logging)

**Request Headers**:

```
Authorization: Bearer <token>
```

**Response** (200 OK):

```json
{
  "message": "Logged out successfully"
}
```

**Example**:

```bash
curl -X POST https://localhost:8443/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## Health Data Endpoints

### 1. Create Health Record

Record a new health measurement.

**Endpoint**: `POST /health`  
**Access**: Protected (requires valid JWT)  
**Security**: CONFIDENTIALITY (user isolation), INTEGRITY (validation), AVAILABILITY (cached)

**Request Body**:

```json
{
  "type": "heart_rate",
  "value": 72.5,
  "unit": "bpm",
  "notes": "After morning coffee",
  "recorded_at": "2025-10-23T08:30:00Z"
}
```

**Validation Rules**:

- `type`: One of: `blood_pressure`, `heart_rate`, `weight`, `temperature`, `glucose` (required)
- `value`: Non-negative number (required)
- `unit`: Unit of measurement (required)
- `notes`: Max 500 characters (optional)
- `recorded_at`: ISO 8601 format (optional, defaults to now)

**Response** (201 Created):

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "heart_rate",
  "value": 72.5,
  "unit": "bpm",
  "notes": "After morning coffee",
  "recorded_at": "2025-10-23T08:30:00Z",
  "created_at": "2025-10-23T08:30:45Z"
}
```

**Error Responses**:

- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Missing/invalid token
- `413 Payload Too Large`: Request exceeds max size (10MB)

**Example**:

```bash
curl -X POST https://localhost:8443/api/v1/health \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "heart_rate",
    "value": 72.5,
    "unit": "bpm",
    "notes": "After morning coffee"
  }'
```

---

### 2. Get Health Records

Retrieve health records for the authenticated user.

**Endpoint**: `GET /health`  
**Access**: Protected (requires valid JWT)  
**Security**: CONFIDENTIALITY (user isolation), AVAILABILITY (cached retrieval)

**Query Parameters**:

- `limit`: Max records to return (default: 20, max: 100)

**Response** (200 OK):

```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "heart_rate",
    "value": 72.5,
    "unit": "bpm",
    "notes": "After morning coffee",
    "recorded_at": "2025-10-23T08:30:00Z",
    "created_at": "2025-10-23T08:30:45Z"
  }
]
```

**Response Headers**:

- `X-Total-Count`: Total number of records returned

**Example**:

```bash
curl -X GET "https://localhost:8443/api/v1/health?limit=50" \
  -H "Authorization: Bearer <token>"
```

---

### 3. Get Health Statistics

Get aggregated statistics for a specific health metric type.

**Endpoint**: `GET /health/stats`  
**Access**: Protected (requires valid JWT)  
**Security**: AVAILABILITY (cached aggregation, 1-hour TTL)

**Query Parameters**:

- `type`: Health metric type (required) - One of: `blood_pressure`, `heart_rate`, `weight`, `temperature`, `glucose`

**Response** (200 OK):

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "heart_rate",
  "average": 71.8,
  "min": 60.0,
  "max": 85.5,
  "count": 15,
  "last_record": "2025-10-23T08:30:00Z"
}
```

**Response Headers**:

- `X-Cache`: `HIT` (cached) or `MISS` (computed)

**Example**:

```bash
curl -X GET "https://localhost:8443/api/v1/health/stats?type=heart_rate" \
  -H "Authorization: Bearer <token>"
```

---

### 4. Delete Health Record

Remove a specific health record.

**Endpoint**: `DELETE /health`  
**Access**: Protected (requires valid JWT)  
**Security**: INTEGRITY (only owner can delete)

**Query Parameters**:

- `id`: Record ID to delete (required)

**Response** (200 OK):

```json
{
  "message": "Record deleted successfully"
}
```

**Error Responses**:

- `400 Bad Request`: Missing 'id' parameter
- `401 Unauthorized`: Missing/invalid token
- `500 Internal Server Error`: Delete failed

**Example**:

```bash
curl -X DELETE "https://localhost:8443/api/v1/health?id=660e8400-e29b-41d4-a716-446655440001" \
  -H "Authorization: Bearer <token>"
```

---

## CIA Triad Implementation

### Confidentiality

| Feature              | Implementation                                        |
| -------------------- | ----------------------------------------------------- |
| **Password Hashing** | bcrypt (cost: 12) - passwords never stored plaintext  |
| **JWT Tokens**       | HS256 signing, 1-hour expiry, loaded from environment |
| **HTTPS/TLS**        | TLS 1.2+, enforced redirect from HTTP                 |
| **User Isolation**   | All queries scoped by user_id from JWT                |
| **Header Security**  | X-Content-Type-Options, X-Frame-Options, HSTS         |

### Integrity

| Feature                | Implementation                                   |
| ---------------------- | ------------------------------------------------ |
| **Input Validation**   | Validator/v10 with email, length, enum checks    |
| **Audit Logging**      | All auth actions logged with timestamp, user, IP |
| **CSRF Protection**    | Token framework ready (see SECURITY.md)          |
| **Request Signing**    | Framework ready for future implementation        |
| **Payload Validation** | Max 10MB body size enforced                      |

### Availability

| Feature            | Implementation                                                    |
| ------------------ | ----------------------------------------------------------------- |
| **Caching**        | Redis cache for user records (24hr TTL) and health data (30 days) |
| **Rate Limiting**  | 60 req/min per IP via httprate                                    |
| **Panic Recovery** | Middleware catches errors, returns 500 safely                     |
| **Timeouts**       | Read: 5s, Write: 10s, Idle: 60s                                   |
| **Stats Caching**  | Aggregated stats cached for 1 hour (marked with X-Cache header)   |

---

## Performance Enhancements

### Caching Strategy

1. **User Cache** (24 hours):

   - Key: `user:<email>`
   - Content: Full user record with hashed password
   - Invalidated on profile update

2. **Health Record Cache** (30 days):

   - Key: `health:<user_id>:<record_id>`
   - Content: Individual health measurements

3. **Health Index** (30 days):

   - Key: `health:<user_id>:list`
   - Content: List of record IDs for pagination

4. **Stats Cache** (1 hour):
   - Key: `health:<user_id>:stats:<type>`
   - Content: Aggregated (avg, min, max, count)
   - Invalidated on new record creation

### Query Optimization

- **Pagination**: Limit 20 by default, max 100 to prevent DoS
- **Lazy Loading**: Stats computed on-demand, cached for reuse
- **List Indexing**: Redis LRANGE for efficient pagination
- **Invalidation**: Smart cache invalidation on writes

### Response Optimization

- **Gzip Compression**: All responses compressed if client supports
- **Cache Headers**: X-Cache header shows hit/miss for debugging
- **Partial Responses**: Only essential fields returned (passwords excluded)

---

## Error Handling

### HTTP Status Codes

| Code | Meaning                                                      |
| ---- | ------------------------------------------------------------ |
| 200  | OK - Request succeeded                                       |
| 201  | Created - Resource created successfully                      |
| 400  | Bad Request - Invalid input or missing required fields       |
| 401  | Unauthorized - Invalid/missing JWT token                     |
| 403  | Forbidden - User inactive or permission denied               |
| 404  | Not Found - Resource not found                               |
| 409  | Conflict - Resource already exists (e.g., email)             |
| 413  | Payload Too Large - Request exceeds max size                 |
| 429  | Too Many Requests - Rate limit exceeded                      |
| 500  | Internal Server Error - Server error (safe message returned) |

### Error Response Format

```json
{
  "error": "Description of what went wrong"
}
```

---

## Rate Limiting

- **Limit**: 60 requests per minute per IP
- **Header**: Returns `429 Too Many Requests` when exceeded
- **Reset**: Automatic after 1 minute

---

## Token Format

JWT tokens include:

- `sub` (subject): User ID
- `exp` (expiration): Token expiry time (1 hour)
- Algorithm: HS256

**Token Lifecycle**:

1. Issued on `/auth/register` or `/auth/login`
2. Included in `Authorization: Bearer <token>` header for protected endpoints
3. Verified on every protected request
4. Expires after 1 hour
5. Client must login again to get new token

---

## Testing with cURL

### Register & Get Token

```bash
TOKEN=$(curl -s -X POST https://localhost:8443/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "full_name": "Test User"
  }' | jq -r '.token')

echo $TOKEN  # Save for next requests
```

### Create Health Record

```bash
curl -X POST https://localhost:8443/api/v1/health \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "heart_rate",
    "value": 72.5,
    "unit": "bpm"
  }'
```

### Get Statistics (with Cache Header)

```bash
curl -X GET "https://localhost:8443/api/v1/health/stats?type=heart_rate" \
  -H "Authorization: Bearer $TOKEN" \
  -i  # Shows response headers including X-Cache
```

---

## Production Deployment

- [ ] Set `JWT_SECRET` environment variable (strong random key)
- [ ] Set `ALLOWED_ORIGINS` for CORS
- [ ] Enable HTTPS with valid certificates (not self-signed)
- [ ] Use external Redis for multi-instance deployments
- [ ] Setup centralized logging (ELK, CloudWatch)
- [ ] Monitor rate limit violations
- [ ] Implement database persistence (PostgreSQL)

---

**API Version**: 1.0  
**Last Updated**: October 23, 2025  
**Status**: Production Ready ✓
