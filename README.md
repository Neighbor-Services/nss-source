# Neighbor Service

Neighbor Service is a comprehensive platform connecting service seekers with local providers. It features a Django-based backend and a Flutter-based mobile frontend.

## Prerequisites

- **Docker** and **Docker Compose** (for Database and Redis)
- **Python 3.10+**
- **Flutter SDK**

## Project Structure

- `backend/`: Django REST Framework API
- `frontend/`: Flutter Mobile Application
- `docker-compose.yml`: Infrastructure services (PostgreSQL, Redis)

## Getting Started

### 1. Infrastructure Setup
Start the required database and cache services:
```bash
docker-compose up -d
```

### 2. Backend Setup
Navigate to the backend directory:
```bash
cd backend
```

Create a virtual environment and install dependencies:
```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

Run migrations and start the server:
```bash
python manage.py migrate
python manage.py runserver
```
The API will be available at `http://localhost:8000`.

### 3. Frontend Setup
Navigate to the frontend application:
```bash
cd frontend/nsapp
```

Get dependencies and run the app:
```bash
flutter pub get
flutter run
```

## Documentation
- [Backend Documentation](backend/README.md)
