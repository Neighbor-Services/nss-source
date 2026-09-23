# Neighbor Services — Golang Backend

A high-performance, concurrent backend built in **Go (Golang)** providing complete API and feature parity with the Django backend (`backend/`), including:

- **JWT Authentication** (Access & Refresh tokens, dual-compatible with Django PBKDF2 and standard bcrypt)
- **Account & Profile Management** (Registration, OTP verification, Password resets, OAuth callbacks)
- **Services & Dynamic Provider Matching** (Haversine geospatial distance calculation + subscription tier ranking boost)
- **Service Requests & Bookings** (Lifecycle status progression, tiered commission calculation: Silver 15%, Gold 10%, Platinum 5%)
- **Reviews & Ratings** (Automated provider rating aggregation)
- **Real-Time WebSockets & Chat** (Gorilla WebSocket duplex rooms, typing indicators, live presence)
- **Push & In-App Notifications** (FCM, APNs device token registration, live notification stream)
- **In-App Purchases (IAP)**:
  - Apple StoreKit 1 (`verifyReceipt`) & StoreKit 2 (`JWS`) validation
  - Google Play Developer API v3 (`androidpublisher`) subscription purchase verification
  - Apple Server-to-Server (S2S) lifecycle webhooks
- **Consultations & Appointments** (Agora RTC/RTM dynamic token generation)
- **Database Compatibility**: GORM models mapped directly to Django database table names (`accounts_user`, `accounts_profile`, `payments_subscription`, etc.) to allow shared database access with zero schema conflicts.

---

## Getting Started

### 1. Environment Configuration

Copy the example environment file:
```bash
cp .env.example .env
```

Configure your `.env` variables (Postgres/SQLite connection, JWT secrets, StoreKit secrets, Google Play Service Account JSON).

### 2. Run Locally

```bash
make run
# Or directly:
go run ./cmd/server
```

The server will start on `http://localhost:8000`.

### 3. Run with Docker

```bash
docker-compose up -d --build
```

### 4. Run Tests

```bash
make test
```
