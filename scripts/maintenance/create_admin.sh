#!/usr/bin/env bash
# ==============================================================================
# Neighbor Service - Administrator Account Creator
# Usage:
#   ./create_admin.sh [email] [password] [firstName] [lastName]
# Example:
#   ./create_admin.sh admin@neighborservice.com Admin123! Super Admin
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend-go"

EMAIL="${1:-admin@neighborservice.com}"
PASSWORD="${2:-Admin123!}"
FIRST_NAME="${3:-Super}"
LAST_NAME="${4:-Admin}"

echo "🔧 Creating / Updating Admin Account..."
echo "📍 Target Email: $EMAIL"

cd "$BACKEND_DIR"

go run ./cmd/create_admin/main.go \
  -email="$EMAIL" \
  -password="$PASSWORD" \
  -first-name="$FIRST_NAME" \
  -last-name="$LAST_NAME"
