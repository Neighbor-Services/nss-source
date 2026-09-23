#!/usr/bin/env bash
# ==============================================================================
# Neighbor Service - Categories & Catalog Services Seeder
# Populates/Syncs 12 Service Categories and 93 Comprehensive Catalog Services.
#
# Usage:
#   ./seed_services.sh
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend-go"

if [ ! -d "$BACKEND_DIR" ]; then
  if [ -d "$SCRIPT_DIR/cmd/seed_services" ]; then
    BACKEND_DIR="$SCRIPT_DIR"
  else
    echo "❌ Error: Could not locate backend-go directory."
    exit 1
  fi
fi

echo "=========================================================="
echo "🌱 Seeding Neighbor Service (NS) Categories & Services..."
echo "=========================================================="

cd "$BACKEND_DIR"

go run ./cmd/seed_services/main.go

echo ""
echo "✅ Seeding script finished successfully!"
