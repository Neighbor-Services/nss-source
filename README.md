# Neighbor Service (NSA)

Neighbor Service is an on-demand service marketplace platform connecting service seekers with verified local service providers.

---

## 🏗 Architecture & Stack Overview

```
                          ┌───────────────────────────┐
                          │   Cloudflare / DNS / SSL  │
                          └─────────────┬─────────────┘
                                        │
           ┌────────────────────────────┼────────────────────────────┐
           │                            │                            │
           ▼                            ▼                            ▼
┌──────────────────────┐     ┌──────────────────────┐     ┌──────────────────────┐
│    Public Website    │     │ Admin Control Center │     │  Flutter Mobile App  │
│ Angular (Port 80/443)│     │ Angular (Port 80/443)│     │  (iOS & Android)     │
└──────────┬───────────┘     └──────────┬───────────┘     └──────────┬───────────┘
           │                            │                            │
           └────────────────────────────┼────────────────────────────┘
                                        │  HTTPS / WSS / REST API
                                        ▼
                          ┌───────────────────────────┐
                          │  Go High-Performance API  │
                          │   (Gin / GORM / Clean)    │
                          └──────┬─────────────┬──────┘
                                 │             │
                    ┌────────────┴─────┐ ┌─────┴────────────┐
                    │ PostgreSQL 16 DB │ │  Redis 7 Cache   │
                    │  (Persistence)   │ │  & Pub/Sub Hub   │
                    └──────────────────┘ └──────────────────┘
```

### Component Summary

| Component | Path | Technology Stack | Description |
| :--- | :--- | :--- | :--- |
| **Backend API** | `backend-go/` | Go 1.26+, Gin, GORM, WebSockets | Core RESTful API, real-time messaging, geo-distance matcher, Stripe, FCM & APNs |
| **Mobile App** | `frontend/nsapp/` | Flutter 3.8+, Dart, BLoC, Dio, Hive | Cross-platform seeker & provider mobile client |
| **Admin Portal** | `frontend/admin-portal/` | Angular 18+, TypeScript, Tailwind/SCSS | Management console for user verification, disputes, and analytics |
| **Public Site** | `frontend/public-site/` | Angular 18+, TypeScript, HTML5 | Marketing landing, terms, resolution center, and support |

---

## 🚀 Quick Start (Local Development)

### 1. Prerequisites
- **Docker** and **Docker Compose**
- **Go 1.22+**
- **Flutter SDK 3.8+**
- **Node.js 20+** & **npm**

---

### 2. Infrastructure Setup
Start the local PostgreSQL and Redis containers:
```bash
docker-compose up -d
```

---

### 3. Backend Setup (Go)
Navigate to the Go backend directory:
```bash
cd backend-go
```

Copy the example environment file:
```bash
cp .env.example .env
```

Run database migrations and start the development server:
```bash
go run ./cmd/server
```

The API and interactive Swagger documentation will be available at:
- **API Base URL**: `http://localhost:8000/api/v1`
- **Health Check**: `http://localhost:8000/health`
- **Interactive Swagger Docs**: `http://localhost:8000/docs`

---

### 4. Mobile App Setup (Flutter)
Navigate to the mobile application directory:
```bash
cd frontend/nsapp
```

Install dependencies:
```bash
flutter pub get
```

Run on connected device or simulator:
```bash
flutter run
```

---

### 5. Web Frontends Setup (Angular)

#### Admin Portal
```bash
cd frontend/admin-portal
npm install
npm start
```
Default URL: `http://localhost:4200`

#### Public Website
```bash
cd frontend/public-site
npm install
npm start
```
Default URL: `http://localhost:4200`

---

## 🧪 Testing

### Backend Unit & Integration Tests
```bash
cd backend-go
go test -v -race ./...
```

### Mobile App Tests
```bash
cd frontend/nsapp
flutter test
```

### Web Frontends Build & Validation
```bash
cd frontend/admin-portal && npm run build
cd frontend/public-site && npm run build
```

---

## 🚢 CI/CD & Production Deployment

### Automated GitHub Actions Workflows
The repository includes automated CI workflows configured under `.github/workflows/`:
- **`backend-ci.yml`**: Go vet, race-condition testing, and binary compilation.
- **`mobile-ci.yml`**: Flutter code analysis and unit/widget testing.
- **`web-ci.yml`**: Angular frontend build verification.

### Production Deployment Scripts
Automated orchestration scripts are available in `scripts/deploy/`:
- **Full Stack Production**: `sudo bash scripts/deploy/deploy-all.sh`
- **Full Stack Staging**: `sudo bash scripts/deploy/deploy-all-staging.sh`
- **Backend API Cluster**: `sudo bash scripts/deploy/deploy-backend.sh`
- **Public Site**: `sudo bash scripts/deploy/deploy-frontend.sh`
- **Admin Portal**: `sudo bash scripts/deploy/deploy-admin-frontend.sh`

---

## 📁 Repository Layout

```
├── .github/workflows/      # Automated CI/CD pipelines
├── backend-go/             # High-performance Go backend service
│   ├── cmd/                # Entrypoints (server, create_admin, seed_services)
│   ├── internal/           # Domain, usecase, repository, delivery, workers
│   └── pkg/                # Reusable utilities (cache, auth, email, media)
├── frontend/
│   ├── nsapp/              # Flutter mobile app (iOS & Android)
│   ├── admin-portal/       # Angular admin dashboard
│   └── public-site/        # Angular public website
├── scripts/
│   ├── deploy/             # Production & staging deployment scripts
│   └── maintenance/        # Operational tasks (seed, restart, admin setup)
├── tools/                  # Asset generators and image tools
├── docs/                   # System guides and presentations
└── docker-compose.yml      # Local development container orchestration
```

---

## 📄 License & Proprietary Notice
© Neighbor Service. All rights reserved.
