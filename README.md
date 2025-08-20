# OTP Server

A high-performance OTP-based authentication service built with Go, designed to handle millions of users per second.

## Features

- **OTP Authentication**: Secure one-time password generation and validation using Redis
- **Rate Limiting**: Built-in rate limiting for OTP requests (max 3 per phone number within 10 minutes)
- **User Management**: RESTful API for user operations with pagination and search
- **High Performance**: Optimized for high-scale operations with Redis caching
- **Containerized**: Docker support with docker-compose
- **Database**: PostgreSQL for user data, Redis for OTP storage and rate limiting

## Architecture

The system follows a clean, layered architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Layer   │    │  Service Layer  │    │ Repository Layer│
│   (Handlers)   │◄──►│   (Services)    │◄──►│  (Repositories) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Middleware    │    │   Domain Layer  │    │   Database      │
│ (Rate Limiting) │    │   (Entities)    │    │  (PostgreSQL)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Redis       │    │   OTP Service   │    │   Rate Limiter  │
│  (OTP Storage)  │    │  (Redis-based)  │    │  (Redis-based)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Database Choice Justification

### Why PostgreSQL for User Data?

**PostgreSQL Advantages:**

1. **ACID Compliance**: Ensures data integrity for critical authentication operations
2. **Performance**: Excellent performance for read-heavy workloads with proper indexing
3. **Scalability**: Can handle millions of users with proper optimization
4. **JSON Support**: Native JSON support for flexible data storage
5. **Partitioning**: Table partitioning for time-series data (OTP expiration)
7. **Extensions**: Rich ecosystem of extensions for additional functionality

**Optimization Features:**
- Efficient phone number lookups with unique constraints
- Simplified schema focused on user management
- Redis handles OTP storage with automatic expiration

### Why Redis for OTP Storage?

**Redis Best Practices for OTPs:**

1. **Automatic Expiration**: Built-in TTL (Time-To-Live) automatically removes expired OTPs
2. **High Performance**: In-memory storage provides microsecond response times
3. **Atomic Operations**: Ensures thread-safe OTP operations
4. **Rate Limiting**: Perfect for implementing request rate limiting
5. **Scalability**: Can handle millions of OTP operations per second
6. **No Cleanup Required**: Unlike database storage, no background cleanup jobs needed

### Alternative Database Considerations

**Why Not MongoDB?**
- **ACID Compliance**: PostgreSQL provides stronger ACID guarantees
- **Schema Validation**: PostgreSQL enforces data structure at the database level
- **Transactions**: Better support for complex transactions
- **Performance**: Better performance for relational data with proper indexing

**Why Not MySQL?**
- **JSON Support**: PostgreSQL has superior JSON handling
- **Extensions**: Richer ecosystem of extensions
- **Performance**: Better performance for complex queries
- **Scalability**: Better horizontal scaling capabilities

## Quick Start

### Prerequisites

- **Docker and Docker Compose** (for containerized setup) - **RECOMMENDED**
- **Go 1.23+** (for local development)
- **PostgreSQL** (for local development)
- **Redis** (for local development)

## How to Run Locally

### Option 1: Using Docker with Makefile (RECOMMENDED)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd otp-server
   ```

2. **Quick Start with Docker (RECOMMENDED) `DO NOT NEED TO CHANGE OR CREATE ANYTHING`**
   ```bash
   make docker-run   # Start all services with Docker

   # Or

   docker compose up
   ```
   
   This single command will:
   - Start PostgreSQL and Redis containers
   - Build and start the OTP server
   - Run database migrations automatically
   - Set up the complete development environment

3. **Alternative: Manual Docker Setup**
   ```bash
   # Check required tools
   make check-tools
   
   # Start all services
   docker-compose up -d
   
   # Run database migrations
   make migrate-up-docker
   ```

### Option 2: Manual Setup with Docker Run Commands

1. **Start PostgreSQL with Docker Run**
   ```bash
   docker run -d --name otp-postgres \
     -e POSTGRES_USER=otp_server_user \
     -e POSTGRES_PASSWORD=otp_server_password \
     -e POSTGRES_DB=otp_server_db \
     -p 5432:5432 \
     postgres:16
   ```

2. **Start Redis with Docker Run (with memory config)**
   ```bash
   docker run -d --name otp-redis \
     -p 6379:6379 \
     redis:7 redis-server --maxmemory 2gb --maxmemory-policy allkeys-lru
   ```

3. **Set environment variables**
   ```bash
   export DB_PROVIDER=postgres
   export POSTGRES_HOST=localhost
   export POSTGRES_PORT=5432
   export POSTGRES_USER=otp_server_user
   export POSTGRES_PASSWORD=otp_server_password
   export POSTGRES_DB=otp_server_db
   export JWT_SECRET=your-secret-key
   export REDIS_HOST=localhost
   export REDIS_PORT=6379
   ```

4. **Run the application**
   ```bash
   go run cmd/main.go
   ```

### Option 3: Local Development Setup

1. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

2. **Set environment variables**
   ```bash
   export DB_PROVIDER=postgres
   export POSTGRES_HOST=localhost
   export POSTGRES_PORT=5432
   export POSTGRES_USER=otp_server_user
   export POSTGRES_PASSWORD=otp_server_password
   export POSTGRES_DB=otp_server_db
   export JWT_SECRET=your-secret-key
   export REDIS_HOST=localhost
   export REDIS_PORT=6379
   ```

3. **Start PostgreSQL and Redis locally**
   ```bash
   # Start PostgreSQL (if using Docker)
   docker run -d --name postgres \
     -e POSTGRES_USER=otp_server_user \
     -e POSTGRES_PASSWORD=otp_server_password \
     -e POSTGRES_DB=otp_server_db \
     -p 5432:5432 \
     postgres:16

   # Start Redis (if using Docker)
   docker run -d --name redis \
     -p 6379:6379 \
     redis:7 redis-server --maxmemory 2gb --maxmemory-policy allkeys-lru
   ```

4. **Run the application**
   ```bash
   go run cmd/main.go
   ```



### Database Migrations with Docker

```bash
# Run database migrations using Docker container (interactive)
make migrate-up-docker

# Run database migrations using Docker container (non-interactive - good for CI/CD)
make migrate-up-docker-ci

# Check database status and tables
make db-status

# Show table structure
make db-schema
```

**Note**: The Docker-based migration commands automatically use the PostgreSQL container credentials and don't require local PostgreSQL installation or `.env` file configuration.

### Docker Commands

```bash
# Build the application image
make docker-build

# Run with Docker Compose
make docker-run

# View logs
make docker-logs

# Stop services
make docker-stop

# Check service status
docker-compose ps

# Rebuild and restart
make docker-stop
make docker-build
make docker-run
```

### Configuration File

A comprehensive configuration template is available at `config.env.example`. Copy this file to `.env` and customize the values for your environment:

```bash
cp config.env.example .env
# Edit .env with your configuration values
```

## Example API Requests & Responses

### Authentication Endpoints

#### 1. Send OTP

**Request:**
```http
POST /api/v1/auth/send-otp
Content-Type: application/json

{
  "phone_number": "+1234567890"
}
```

**Success Response (200):**
```json
{
  "message": "OTP sent successfully",
  "phone_number": "+1234567890"
}
```

**Error Response (400 - Invalid Phone):**
```json
{
  "error": "invalid_phone_number",
  "message": "Invalid phone number format"
}
```

**Error Response (429 - Rate Limited):**
```json
{
  "error": "rate_limit_exceeded",
  "message": "too many requests. Limit: 3 requests per 10m0s. Please try again later."
}
```

**Headers:**
```
Retry-After: 600
X-RateLimit-Exceeded: true
X-RateLimit-Limit: 3
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 600
```

#### 2. Verify OTP and Authenticate

**Request:**
```http
POST /api/v1/auth/verify-otp
Content-Type: application/json

{
  "phone_number": "+1234567890",
  "otp_code": "123456",
  "name": "John Doe"
}
```

**Success Response (200):**
```json
{
  "user": {
    "id": 1,
    "phone_number": "+1234567890",
    "name": "John Doe",
    "role": "user",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Response (400 - Invalid OTP):**
```json
{
  "error": "invalid_otp",
  "message": "invalid OTP code"
}
```

**Error Response (400 - Expired OTP):**
```json
{
  "error": "otp_expired",
  "message": "OTP not found or expired"
}
```

### User Management Endpoints

#### 3. Get Users (with pagination and search)

**Request - List Users:**
```http
GET /api/v1/users/search?offset=0&limit=10
Authorization: Bearer <jwt_token>
```

**Request - Search Users:**
```http
GET /api/v1/users/search?query=john&offset=0&limit=10
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "users": [
    {
      "id": 1,
      "phone_number": "+1234567890",
      "name": "John Doe",
      "role": "user",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "phone_number": "+1234567891",
      "name": "Jane Smith",
      "role": "admin",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 2,
  "offset": 0,
  "limit": 10
}
```

**Error Response (401 - Unauthorized):**
```json
{
  "error": "authorization_required",
  "message": "Please provide a valid authorization header"
}
```

#### 4. Get User Profile

**Request:**
```http
GET /api/v1/users/profile
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "id": 1,
  "phone_number": "+1234567890",
  "name": "John Doe",
  "role": "user",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### 5. Update User Profile

**Request:**
```http
PUT /api/v1/users/profile
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "John Smith"
}
```

**Success Response (200):**
```json
{
  "id": 1,
  "phone_number": "+1234567890",
  "name": "John Smith",
  "role": "user",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Health and Metrics Endpoints

#### 6. Health Check

**Request:**
```http
GET /health
```

**Success Response (200):**
```json
{
  "status": "ok",
  "service": "otp-server",
  "version": "1.0.0"
}
```

#### 7. Metrics (Prometheus)

**Request:**
```http
GET /metrics
```

**Response:**
```prometheus
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="POST",path="/api/v1/auth/send-otp",status_code="200"} 15

# HELP otp_operations_total Total number of OTP operations
# TYPE otp_operations_total counter
otp_operations_total{operation="generate",success="true"} 12

# HELP rate_limit_exceeded_total Total number of rate limit violations
# TYPE rate_limit_exceeded_total counter
rate_limit_exceeded_total{endpoint_type="otp"} 3
```

### Testing with cURL

```bash
# Send OTP
curl -X POST http://localhost:8080/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}'

# Verify OTP (replace 123456 with actual OTP from logs)
curl -X POST http://localhost:8080/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "otp_code": "123456", "name": "John Doe"}'

# Get users (replace TOKEN with JWT from previous response)
curl -X GET "http://localhost:8080/api/v1/users/search?offset=0&limit=10" \
  -H "Authorization: Bearer TOKEN"

# Update profile
curl -X PUT http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "John Smith"}'
```

## Rate Limiting

The system implements multiple levels of rate limiting:

- **Global**: 100 requests per minute per IP
- **Authentication**: 5 requests per 10 minutes per IP
- **OTP Generation**: 3 requests per 10 minutes per phone number
- **User Operations**: 50 requests per minute per IP

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum allowed requests
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time until rate limit resets

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| **Server Configuration** |
| `SERVER_PORT` | 8080 | HTTP server port |
| `SERVER_HOST` | localhost | HTTP server host |
| `ENVIRONMENT` | development | Application environment |
| **Database Configuration** |
| `DB_PROVIDER` | postgres | Database provider |
| `POSTGRES_HOST` | localhost | PostgreSQL host |
| `POSTGRES_PORT` | 5432 | PostgreSQL port |
| `POSTGRES_USER` | otp_server_user | Database username |
| `POSTGRES_PASSWORD` | otp_server_password | Database password |
| `POSTGRES_DB` | otp_server_db | Database name |
| `POSTGRES_SSL_MODE` | disable | PostgreSQL SSL mode |
| `DB_MAX_OPEN_CONNS` | 25 | Maximum open connections |
| `DB_MAX_IDLE_CONNS` | 5 | Maximum idle connections |
| `DB_CONN_MAX_LIFETIME` | 1h | Connection max lifetime |
| **Redis Configuration** |
| `REDIS_HOST` | localhost | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `REDIS_PASSWORD` | - | Redis password |
| `REDIS_DB` | 0 | Redis database number |
| `REDIS_POOL_SIZE` | 10 | Redis connection pool size |
| `REDIS_MIN_IDLE_CONNS` | 5 | Redis minimum idle connections |
| `REDIS_MAX_RETRIES` | 3 | Redis max retry attempts |
| `REDIS_CLUSTER_MODE` | false | Enable Redis cluster mode |
| **JWT Configuration** |
| `JWT_SECRET` | - | JWT signing secret |
| `JWT_EXPIRY` | 24000h | JWT token expiry |
| `JWT_REFRESH_EXPIRY` | 42000h | JWT refresh token expiry |
| **Logging Configuration** |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | json | Log format (json, text) |
| `LOG_OUTPUT` | stdout | Log output destination |
| `LOG_FILE_PATH` | ./logs/app.log | Log file path |
| `LOG_MAX_SIZE` | 100 | Maximum log file size (MB) |
| `LOG_MAX_BACKUPS` | 3 | Maximum log file backups |
| `LOG_MAX_AGE` | 28 | Maximum log file age (days) |
| **CORS Configuration** |
| `CORS_ALLOWED_ORIGINS` | http://localhost:3000,http://localhost:8080 | Allowed CORS origins |
| `CORS_ALLOWED_METHODS` | GET,POST,PUT,DELETE,OPTIONS | Allowed HTTP methods |
| `CORS_ALLOWED_HEADERS` | Content-Type,Authorization,X-Requested-With | Allowed HTTP headers |
| **Metrics Configuration** |
| `METRICS_ENABLED` | true | Enable metrics collection |
| `METRICS_PROVIDER` | prometheus | Metrics provider |
| `METRICS_ENDPOINT` | /metrics | Metrics endpoint |
| **OTP Configuration** |
| `OTP_EXPIRY` | 2m | OTP expiration time |
| `OTP_LENGTH` | 6 | OTP code length |
| `OTP_REDIS_KEY_PREFIX` | otp | Redis key prefix for OTPs |
| `OTP_CODE_CHARSET` | 0123456789 | Characters used for OTP generation |
| **Events Configuration** |
| `EVENTS_ENABLED` | true | Enable event system |
| `EVENTS_REDIS_CHANNEL` | events | Redis channel for events |
| `EVENTS_BATCH_SIZE` | 100 | Event batch size |
| `EVENTS_FLUSH_INTERVAL` | 5s | Event flush interval |
| **Rate Limiting Configuration** |
| `RATE_LIMIT_GLOBAL_REQUESTS` | 100 | Global rate limit requests |
| `RATE_LIMIT_GLOBAL_DURATION` | 1m | Global rate limit duration |
| `RATE_LIMIT_AUTH_REQUESTS` | 20 | Auth rate limit requests |
| `RATE_LIMIT_AUTH_DURATION` | 1m | Auth rate limit duration |
| `RATE_LIMIT_OTP_REQUESTS` | 3 | OTP rate limit requests |
| `RATE_LIMIT_OTP_DURATION` | 10m | OTP rate limit duration |
| `RATE_LIMIT_USER_REQUESTS` | 50 | User operations rate limit |
| `RATE_LIMIT_USER_DURATION` | 1m | User operations rate limit duration |

## Development

### Project Structure

```
otp-server/
├── cmd/                    # Application entry point
├── internal/              # Internal application code
│   ├── application/       # Business logic services
│   ├── domain/           # Domain entities and interfaces
│   ├── infrastructure/   # External concerns (DB, Redis, etc.)
│   └── interfaces/       # HTTP handlers and middleware
├── migrations/            # Database migrations
├── docker-compose.yml     # Docker services configuration
├── Dockerfile            # Application container
└── README.md             # This file
```

### Adding New Features

1. **Domain Layer**: Add entities and repository interfaces
2. **Repository Layer**: Implement data access logic
3. **Service Layer**: Add business logic
4. **Handler Layer**: Create HTTP endpoints
5. **Middleware**: Add any required middleware

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/application/services
```



## Performance Considerations

### Database Optimization

- **Indexing**: Strategic indexes on frequently queried columns
- **Partitioning**: OTP table can be partitioned by date for better performance
- **Cleanup**: Automatic cleanup of expired OTPs
- **Connection Pooling**: Efficient database connection management

### Caching Strategy

- **Redis**: Used for rate limiting and session management
- **In-Memory**: Local caching for frequently accessed data
- **TTL**: Automatic expiration for temporary data

### Pub/Sub (Publish/Subscribe) System for Authentication

The OTP server implements a Redis-based pub/sub system specifically designed for authentication events and real-time communication. This system is crucial for avoiding synchronous programming patterns and handling millions of users efficiently:

**Key Features:**
- **Authentication Event Broadcasting**: Real-time notification of all authentication-related events
- **Decoupled Architecture**: Authentication services communicate asynchronously without direct coupling
- **Scalable Messaging**: Multiple subscribers can receive the same authentication events
- **Redis Channels**: Uses Redis pub/sub channels for reliable message delivery
- **Async Processing**: Eliminates blocking operations for OTP generation and verification
- **High Throughput**: Designed to handle millions of concurrent OTP operations

**Why Pub/Sub for OTP?**
- **Avoid Sync Programming**: Eliminates blocking calls that could slow down the system under high load
- **Handle Millions of Users**: Asynchronous processing allows the system to scale horizontally
- **Non-blocking Operations**: OTP generation and verification happen in the background
- **Real-time Notifications**: Immediate event broadcasting for security monitoring
- **Load Distribution**: Multiple workers can process OTP events concurrently

**Authentication Event Types:**
- `otp:generated`: When a new OTP is created for authentication
- `otp:verified`: When an OTP is successfully verified and user is authenticated
- `otp:expired`: When an OTP expires (failed authentication attempt)
- `user:authenticated`: When a user successfully authenticates with OTP
- `user:created`: When a new user is created during authentication
- `auth:failed`: When authentication attempts fail (invalid OTP, expired OTP)
- `auth:rate_limited`: When authentication requests are rate limited

**Authentication Use Cases:**
- **Real-time Notifications**: Send push notifications when OTPs are generated for authentication
- **Audit Logging**: Log all authentication events for security compliance and monitoring
- **Analytics**: Track authentication success/failure rates and OTP usage patterns
- **Security Monitoring**: Real-time alerts for suspicious authentication patterns
- **Integration**: Connect with external security systems (SIEM, fraud detection platforms)
- **Multi-factor Authentication**: Coordinate with additional authentication factors
- **High-Scale Operations**: Handle millions of concurrent OTP requests without blocking

**Example Event Structure:**
```json
{
  "event_type": "otp:generated",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "phone_number": "+1234567890",
    "otp_code": "123456",
    "expires_at": "2024-01-01T12:02:00Z"
  }
}
```

### Cache Query System for User Data

The system implements an intelligent caching layer specifically optimized for user-related data and frequently accessed user information:

**Cache Layers:**
1. **L1 Cache (In-Memory)**: Fastest access for hot user data
2. **L2 Cache (Redis)**: Distributed cache for shared user data across instances
3. **L3 Cache (Database)**: Persistent storage for all user data

**Cache Query Strategy:**
- **Cache-Aside Pattern**: Check cache first, then database for user data
- **Write-Through**: Update cache when writing user data to database
- **TTL Management**: Automatic expiration based on user data freshness requirements
- **Cache Invalidation**: Smart invalidation when user data is updated

**User Data Caching:**
- **User Profiles**: Frequently accessed user information (name, phone, role, status)
- **User Sessions**: Active user sessions and authentication tokens
- **User Permissions**: Cached user roles and access rights
- **User Search Results**: Cached search results for user queries
- **User Rate Limits**: Current rate limit state per user/IP for user operations

**User Data Cache Query Flow:**
```
User Request → Check L1 Cache (User Data) → Check L2 Cache (Shared User Data) → Query Database (User Table) → Update Caches → Return User Response
```

**User Data Performance Benefits:**
- **Response Time**: 90%+ of user profile requests served from cache (sub-1ms response)
- **Database Load**: Significant reduction in user-related database queries
- **Scalability**: Horizontal scaling with shared user data cache across instances
- **Availability**: Graceful degradation when cache is unavailable, fallback to database
- **User Experience**: Instant user profile loading and search results

### Scalability Features

- **Horizontal Scaling**: Stateless application design
- **Load Balancing**: Ready for load balancer deployment
- **Database Sharding**: Schema supports horizontal partitioning
- **Microservices**: Modular design for service decomposition

## Monitoring and Health Checks

- **Health Endpoint**: `/health` for service health monitoring
- **Metrics**: Prometheus metrics endpoint at `/metrics`
- **Logging**: Structured logging with configurable levels
- **Tracing**: OpenTelemetry support for distributed tracing

## Security Features

- **JWT Authentication**: Secure token-based authentication
- **Rate Limiting**: Protection against abuse and DoS attacks
- **Input Validation**: Comprehensive input sanitization
- **SQL Injection Protection**: Parameterized queries
- **HTTPS Ready**: TLS support for production deployment

## Production Deployment

### Environment Setup

1. **Set strong JWT secret**
2. **Configure production database**
3. **Set up monitoring and alerting**
4. **Configure backup strategies**

### Scaling Considerations

- **Database**: Use read replicas for read-heavy workloads
- **Application**: Deploy multiple instances behind load balancer
- **Caching**: Redis cluster for high availability
- **Monitoring**: Comprehensive observability stack

## Troubleshooting

### Common Issues

1. **Database Connection**: Check PostgreSQL service status
2. **Rate Limiting**: Verify Redis connection
3. **OTP Expiration**: Check system time synchronization
4. **Performance**: Monitor database query performance

### Logs

Application logs are written to stdout/stderr and can be viewed with:
```bash
docker-compose logs -f otp-server
```

## Future Work & Scaling Roadmap

### Database Scaling

### Redis High Availability & Memory Management
- **Redis Cluster**: Implement Redis cluster mode for horizontal scaling
- **Memory Configuration**: Advanced memory management to prevent memory overflow
- **Garbage Collector Configuration**: Optimize Redis memory cleanup and eviction policies
- **Memory Monitoring**: Real-time memory usage tracking and alerting
- **Backup & Recovery**: Automated Redis backup strategies and disaster recovery
- **Memory Policies**: Configurable eviction policies (LRU, LFU, TTL-based)
- **Memory Limits**: Set maximum memory limits with graceful degradation



### Application High Availability
- **Load Balancer Integration**: Deploy behind HAProxy, Nginx, or cloud load balancers
- **Auto-scaling**: Implement horizontal pod autoscaling (HPA) for Kubernetes

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request


## Support

For support and questions:
- Create an issue in the repository
- Check the documentation
- Review the API examples

---

**Note**: This is a development version. For production use, ensure proper security measures, monitoring, and backup strategies are in place. 