# OTP Server API Documentation

## Overview

The OTP Server provides a secure, scalable API for OTP-based authentication and user management. This document describes all available endpoints, request/response formats, and usage examples.

## Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://your-domain.com`

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the `Authorization` header:

```
Authorization: Bearer <your_jwt_token>
```

## Rate Limiting

- **OTP Requests**: Maximum 3 requests per phone number within 10 minutes
- **API Endpoints**: Standard rate limiting applied to all endpoints

## API Endpoints

### 1. Authentication

#### Send OTP

Sends a 6-digit OTP to the specified phone number.

```http
POST /api/v1/auth/send-otp
Content-Type: application/json
```

**Request Body:**
```json
{
  "phone_number": "+1234567890"
}
```

**Response (200 OK):**
```json
{
  "message": "OTP sent successfully",
  "phone_number": "+1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid phone number format
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

**Notes:**
- OTP is printed to console for development
- OTP expires after 2 minutes
- Rate limited to 3 requests per phone number per 10 minutes

#### Verify OTP

Verifies the OTP and returns authentication tokens.

```http
POST /api/v1/auth/verify-otp
Content-Type: application/json
```

**Request Body:**
```json
{
  "phone_number": "+1234567890",
  "otp": "123456"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "phone_number": "+1234567890",
    "name": "",
    "role": "user"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Invalid or expired OTP
- `500 Internal Server Error`: Server error

**Notes:**
- New users are automatically created upon first verification
- Access token is valid for 24 hours
- Refresh token is valid for 7 days

#### Refresh Token

Refreshes an expired access token using a valid refresh token.

```http
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Token refreshed successfully",
  "expires_in": 3600
}
```

**Error Responses:**
- `400 Bad Request`: Missing or invalid authorization header
- `401 Unauthorized`: Invalid or expired refresh token
- `500 Internal Server Error`: Server error

### 2. User Management

#### Get User Profile

Retrieves the current user's profile information.

```http
GET /api/v1/users/profile
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "id": 1,
  "phone_number": "+1234567890",
  "name": "John Doe",
  "role": "user",
  "is_active": true,
  "last_seen": "2024-01-15T10:30:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid or expired token
- `404 Not Found`: User not found
- `500 Internal Server Error`: Server error

#### Update User Profile

Updates the current user's profile information.

```http
PUT /api/v1/users/profile
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "John Doe",
  "phone_number": "+1234567890"
}
```

**Response (200 OK):**
```json
{
  "id": 1,
  "phone_number": "+1234567890",
  "name": "John Doe",
  "role": "user",
  "is_active": true,
  "last_seen": "2024-01-15T10:30:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Invalid or expired token
- `409 Conflict`: Phone number already exists
- `500 Internal Server Error`: Server error

#### Get Users (Admin Only)

Retrieves a paginated list of all users. Admin access required.

```http
GET /api/v1/users?page=1&limit=10
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": 1,
      "phone_number": "+1234567890",
      "name": "John Doe",
      "role": "user",
      "is_active": true,
      "last_seen": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: Insufficient permissions
- `500 Internal Server Error`: Server error

#### Search Users (Admin Only)

Searches for users by name or phone number. Admin access required.

```http
GET /api/v1/users/search?q=john
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `q` (required): Search query string

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": 1,
      "phone_number": "+1234567890",
      "name": "John Doe",
      "role": "user",
      "is_active": true,
      "last_seen": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Error Responses:**
- `400 Bad Request`: Missing search query
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: Insufficient permissions
- `500 Internal Server Error`: Server error

### 3. System Endpoints

#### Health Check

Checks the health status of the service.

```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s"
}
```

#### Metrics

Returns Prometheus metrics for monitoring.

```http
GET /metrics
```

**Response (200 OK):**
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="POST",endpoint="/api/v1/auth/send-otp"} 42
```

## Error Handling

All error responses follow a consistent format:

```json
{
  "error": "Error Type",
  "message": "Detailed error message",
  "timestamp": "2024-01-15T10:30:00Z",
  "path": "/api/v1/auth/send-otp"
}
```

### Common Error Types

- `ValidationError`: Input validation failed
- `AuthenticationError`: Authentication failed
- `AuthorizationError`: Insufficient permissions
- `NotFoundError`: Resource not found
- `RateLimitError`: Rate limit exceeded
- `InternalError`: Server internal error

## Data Models

### User

```json
{
  "id": 1,
  "phone_number": "+1234567890",
  "name": "John Doe",
  "role": "user",
  "is_active": true,
  "last_seen": "2024-01-15T10:30:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

**Fields:**
- `id`: Unique user identifier (auto-increment)
- `phone_number`: User's phone number (unique, international format)
- `name`: User's display name
- `role`: User role (`user` or `admin`)
- `is_active`: Whether the account is active
- `last_seen`: Last activity timestamp
- `created_at`: Account creation timestamp
- `updated_at`: Last profile update timestamp

### Authentication Response

```json
{
  "access_token": "JWT token for API access",
  "refresh_token": "JWT token for refreshing access token",
  "user": {
    "id": 1,
    "phone_number": "+1234567890",
    "name": "John Doe",
    "role": "user"
  }
}
```

## Usage Examples

### Complete Authentication Flow

1. **Send OTP**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/send-otp \
     -H "Content-Type: application/json" \
     -d '{"phone_number": "+1234567890"}'
   ```

2. **Verify OTP** (check console for OTP)
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/verify-otp \
     -H "Content-Type: application/json" \
     -d '{"phone_number": "+1234567890", "otp": "123456"}'
   ```

3. **Use Access Token**
   ```bash
   curl -X GET http://localhost:8080/api/v1/users/profile \
     -H "Authorization: Bearer <access_token>"
   ```

### Admin Operations

1. **Get All Users**
   ```bash
   curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10" \
     -H "Authorization: Bearer <admin_access_token>"
   ```

2. **Search Users**
   ```bash
   curl -X GET "http://localhost:8080/api/v1/users/search?q=john" \
     -H "Authorization: Bearer <admin_access_token>"
   ```

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **OTP Requests**: 3 requests per phone number per 10 minutes
- **General API**: Standard rate limiting applied to all endpoints
- **Headers**: Rate limit information included in response headers

## Security Considerations

1. **JWT Tokens**: Store securely and never expose in client-side code
2. **Phone Numbers**: Validate format and consider international standards
3. **Rate Limiting**: Implement client-side retry logic with exponential backoff
4. **HTTPS**: Always use HTTPS in production
5. **Token Expiration**: Implement automatic token refresh logic

## Testing

### Test Environment

- Use the development environment for testing
- OTPs are printed to console (no SMS costs)
- Test data is automatically cleaned up

### Test Data

- Default admin user: `+1234567890` / `Admin User`
- Create test users with different phone numbers
- Test rate limiting with multiple requests

## Support

For API support and questions:
- Check the service logs for detailed error information
- Verify request format and authentication headers
- Ensure rate limits are not exceeded
- Contact the development team for additional assistance 