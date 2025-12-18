# FitSlot - Gym Booking Backend

A production-grade Go backend application for managing gym class bookings with JWT authentication, role-based access control, and comprehensive features.

## Features

- **User Management**: Registration, login, JWT-based authentication with refresh tokens
- **Gym & Time Slot Management**: Create and manage gyms with time slots
- **Booking System**: Book, cancel, and view bookings with subscription and wallet payment support
- **Payment Integration**: Wallet system and subscription plans
- **Email Notifications**: Background worker for sending booking confirmation emails
- **Rate Limiting**: In-memory rate limiter to prevent abuse
- **Structured Logging**: JSON-based structured logging using slog
- **Metrics**: Prometheus metrics for monitoring
- **Graceful Shutdown**: Proper cleanup on application termination
- **Request Validation**: Middleware for validating request bodies
- **API Documentation**: Swagger/OpenAPI documentation
- **Unit & Integration Tests**: Comprehensive test coverage

## Architecture

The application follows the **Standard Go Project Layout** with clean architecture principles:

```
fitslot/
├── cmd/app/              # Application entry point
├── internal/
│   ├── auth/            # Authentication & authorization
│   ├── booking/         # Booking domain (handler, service, repository, model)
│   ├── config/          # Configuration management
│   ├── db/              # Database connection & migrations
│   ├── email/           # Email service with background worker
│   ├── gym/             # Gym domain
│   ├── logger/          # Structured logging
│   ├── metrics/         # Prometheus metrics
│   ├── server/          # HTTP server setup & middleware
│   ├── subscription/    # Subscription domain
│   ├── user/            # User domain
│   └── wallet/          # Wallet domain
├── migrations/          # Database migrations (golang-migrate)
└── docker-compose.yml   # Docker Compose setup
```

### Layer Architecture

- **Handler**: HTTP request/response handling (Gin framework)
- **Service**: Business logic layer with interfaces
- **Repository**: Data access layer with interfaces
- **Model**: Domain models

## Prerequisites

- Go 1.23+
- PostgreSQL 15+
- Redis (for email queue)
- Docker & Docker Compose (optional)

## Setup

### 1. Clone the repository

```bash
git clone <repository-url>
cd fitslot
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure environment variables

Create a `.env` file:

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/fitslot?sslmode=disable
JWT_SECRET=your-secret-key-change-in-production
EMAIL_FROM=noreply@fitslot.com
EMAIL_FROM_NAME=FitSlot
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
REDIS_ADDR=localhost:6379
```

### 4. Run with Docker Compose

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database
- Redis
- MailHog (for email testing)
- Prometheus
- The application

### 5. Run migrations

Migrations run automatically on application startup using `golang-migrate`.

To manually seed the database:

```bash
psql $DATABASE_URL -f migrations/seed.sql
```

## API Endpoints

### Authentication

#### Register
```http
POST /auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "member"
  }
}
```

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

#### Refresh Token
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJ..."
}
```

### User

#### Get Current User
```http
GET /me
Authorization: Bearer <access_token>
```

### Gyms

#### List Gyms
```http
GET /gyms
Authorization: Bearer <access_token>
```

#### List Time Slots
```http
GET /gyms/:gymID/slots
Authorization: Bearer <access_token>
```

### Bookings

#### Book a Slot
```http
POST /slots/:slotID/book
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "booking": {
    "id": 1,
    "user_id": 1,
    "time_slot_id": 1,
    "status": "booked",
    "created_at": "2024-01-15T10:00:00Z"
  },
  "paid_with": "wallet",
  "amount_cents": 1000
}
```

#### Cancel Booking
```http
POST /bookings/:bookingID/cancel
Authorization: Bearer <access_token>
```

#### List My Bookings
```http
GET /bookings
Authorization: Bearer <access_token>
```

### Wallet

#### Get Balance
```http
GET /wallet
Authorization: Bearer <access_token>
```

#### Top Up Wallet
```http
POST /wallet/topup
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "amount_cents": 10000
}
```

#### List Transactions
```http
GET /wallet/transactions?limit=50&offset=0
Authorization: Bearer <access_token>
```

### Subscriptions

#### Create Subscription
```http
POST /subscriptions
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "type": "single_gym_lite",
  "gym_id": 1
}
```

#### List My Subscriptions
```http
GET /subscriptions
Authorization: Bearer <access_token>
```

#### List Subscription Plans
```http
GET /subscriptions/plans
Authorization: Bearer <access_token>
```

### Admin Endpoints

All admin endpoints require `admin` role.

#### Create Gym
```http
POST /admin/gyms
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "New Gym",
  "location": "123 Main St"
}
```

#### Create Time Slot
```http
POST /admin/gyms/:gymID/slots
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "start_time": "2024-01-20T10:00:00Z",
  "end_time": "2024-01-20T11:00:00Z",
  "capacity": 20
}
```

#### List Bookings by Slot
```http
GET /admin/slots/:slotID/bookings
Authorization: Bearer <access_token>
```

#### List Bookings by Gym
```http
GET /admin/gyms/:gymID/bookings
Authorization: Bearer <access_token>
```

## Testing

### Run Unit Tests

```bash
go test ./...
```

Or use the Makefile:

```bash
make test
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Integration Tests

Integration tests are located in the `integration/` directory.

**Option 1: Using Docker Compose (Recommended)**

Start the test database:
```bash
docker-compose -f docker-compose.test.yml up -d
```

Run integration tests:
```bash
go test ./integration/... -v
```

**Option 2: Using Environment Variables**

Set environment variables for test database connection:
```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=testuser
export TEST_DB_PASSWORD=password
export TEST_DB_NAME=fitslot_test
```

Or use a full DSN:
```bash
export TEST_DSN=postgres://testuser:password@localhost:5433/fitslot_test?sslmode=disable
```

Then run tests:
```bash
go test ./integration/... -v
```

**Note:** Integration tests will skip if they cannot connect to the test database.

## API Documentation

### Swagger/OpenAPI

The API is documented using Swagger/OpenAPI. To generate and view the documentation:

1. Install Swagger CLI:
```bash
make install-swagger
# or
go install github.com/swaggo/swag/cmd/swag@latest
```

2. Generate documentation:
```bash
make swagger
# or
swag init -g cmd/app/main.go -o docs
```

3. Start the server and visit:
```
http://localhost:8080/swagger/index.html
```

The Swagger UI provides interactive API documentation where you can test endpoints directly.

## Development

### Project Structure

The codebase follows clean architecture principles:

- **Interfaces**: All layers use interfaces for dependency injection
- **Context Propagation**: All database operations use `context.Context`
- **Error Handling**: Proper error wrapping and handling throughout
- **No Panics**: All errors are handled gracefully
- **Service Layer**: Business logic is separated in service layer
- **Dependency Injection**: All dependencies are injected through constructors

### Adding a New Feature

1. Create model in `internal/<domain>/model.go`
2. Create repository interface in `internal/<domain>/repository_interface.go`
3. Implement repository in `internal/<domain>/repository.go`
4. Create service interface in `internal/<domain>/service.go`
5. Implement service in `internal/<domain>/service.go`
6. Create handler that uses the service
7. Register routes in `internal/server/server.go`
8. Add unit tests for the service layer
9. Update Swagger documentation

### Makefile Commands

- `make swagger` - Generate Swagger documentation
- `make test` - Run unit tests
- `make integration-test` - Run integration tests
- `make build` - Build the application
- `make run` - Run the application
- `make deps` - Download and tidy dependencies

## Observability

### Logging

Structured JSON logging is used throughout the application. Logs include:
- Request method, path, status, latency
- Client IP and user agent
- Error details with context

### Metrics

Prometheus metrics are exposed at `/metrics`:

- `fitslot_http_requests_total`: Total HTTP requests
- `fitslot_http_request_duration_seconds`: Request latency
- `fitslot_bookings_total`: Booking statistics
- `fitslot_emails_sent_total`: Email statistics

### Health Check

```http
GET /health
```

## Security

- **JWT Authentication**: Access tokens (15 min) and refresh tokens (7 days)
- **Password Hashing**: bcrypt with default cost
- **Rate Limiting**: 100 requests/second per IP
- **CORS**: Configurable CORS middleware
- **Role-Based Access Control**: Admin and member roles

## Background Workers

### Email Service

The email service runs as a background worker using goroutines and channels:
- Processes email queue from Redis
- Retries failed emails up to 3 times
- Sends booking confirmations, reminders, and cancellations

## Database Migrations

Migrations are managed using `golang-migrate`:

```bash
# Create a new migration
migrate create -ext sql -dir migrations -seq <migration_name>

# Run migrations (automatic on startup)
# Or manually:
migrate -path migrations -database $DATABASE_URL up
```

## Deployment

### Docker

```bash
docker build -t fitslot .
docker run -p 8080:8080 --env-file .env fitslot
```

### Environment Variables

Required environment variables:
- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret for JWT signing
- `REDIS_ADDR`: Redis address for email queue
- SMTP configuration for email sending


