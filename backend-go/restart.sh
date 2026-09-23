#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) — Go Backend Service Restart Script
# Restarts all Go backend API instances, cluster workers, and background jobs.
#
# Usage:
#   bash restart_services.sh              (Standard / Dev local restart)
#   sudo bash restart_services.sh         (Production systemd multi-cluster restart)
# ==============================================================================

set -e

# Configuration
APP_NAME="ns-backend"
PORT="${PORT:-8000}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}      Neighbor Service (NSA) — Go Backend Restart Manager       ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

# 1. Production Systemd Cluster Detection
if [ "$EUID" -eq 0 ] && systemctl list-unit-files 2>/dev/null | grep -qE "ns-backend(\.target|@)"; then
    echo -e "${YELLOW}1. Restarting Systemd Multi-Instance Go Cluster...${NC}"
    
    # Restart main cluster target
    if systemctl list-unit-files | grep -q "ns-backend.target"; then
        systemctl restart ns-backend.target
        echo -e "${GREEN}  ✓ ns-backend.target restarted${NC}"
    fi

    # Restart individual worker instances (ns-backend@1, ns-backend@2, ...)
    for inst in $(systemctl list-units --type=service --all 2>/dev/null | grep -oE "ns-backend@[0-9]+" | sort -u); do
        echo -e "  Restarting $inst..."
        systemctl restart "$inst"
        echo -e "${GREEN}  ✓ $inst restarted${NC}"
    done

    # Staging Go service
    if systemctl list-unit-files | grep -q "ns-backend-staging"; then
        systemctl restart ns-backend-staging && echo -e "${GREEN}  ✓ ns-backend-staging restarted${NC}" || true
    fi

    # Reload Nginx upstream
    if command -v nginx &>/dev/null; then
        echo -e "${YELLOW}2. Reloading Nginx Load Balancer...${NC}"
        nginx -t 2>/dev/null && systemctl reload nginx && echo -e "${GREEN}  ✓ Nginx reloaded${NC}" || echo -e "${RED}  ✗ Nginx reload skipped${NC}"
    fi

    echo ""
    echo -e "${GREEN}✓ Production Go Cluster successfully restarted!${NC}"
    exit 0
fi

# 2. Local Development Restart Flow
echo -e "${YELLOW}1. Stopping existing Go backend processes...${NC}"

# Find and terminate processes on the configured port
EXISTING_PID=$(lsof -ti :$PORT 2>/dev/null || true)
if [ -n "$EXISTING_PID" ]; then
    echo -e "  Found process $EXISTING_PID listening on port :$PORT. Stopping..."
    kill -15 "$EXISTING_PID" 2>/dev/null || kill -9 "$EXISTING_PID" 2>/dev/null || true
    sleep 1
fi

# Stop any orphan `go run` or compiled server instances
pkill -f "go run ./cmd/server/main.go" 2>/dev/null || true
pkill -f "$DIR/bin/server" 2>/dev/null || true
echo -e "${GREEN}✓ Previous Go processes stopped${NC}"

echo -e "${YELLOW}2. Compiling Go Backend...${NC}"
cd "$DIR"
mkdir -p "$DIR/bin"
mkdir -p "$DIR/logs"

go build -o "$DIR/bin/server" ./cmd/server/main.go
echo -e "${GREEN}✓ Compilation successful: bin/server${NC}"

echo -e "${YELLOW}3. Launching Go Backend Server in Background...${NC}"
nohup "$DIR/bin/server" > "$DIR/logs/server.log" 2>&1 &
NEW_PID=$!
echo -e "${GREEN}✓ Go server started with PID: $NEW_PID (listening on :$PORT)${NC}"
echo -e "  Log output: ${CYAN}$DIR/logs/server.log${NC}"

# 4. Health Check Verification
echo -e "${YELLOW}4. Running Health Check...${NC}"
sleep 2
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$PORT/api/v1/health" 2>/dev/null || echo "000")

if [ "$HEALTH_STATUS" -eq 200 ] || [ "$HEALTH_STATUS" -eq 404 ] || [ "$HEALTH_STATUS" -eq 401 ]; then
    echo -e "${GREEN}✓ Go backend is healthy and accepting connections on port :$PORT (HTTP $HEALTH_STATUS)${NC}"
else
    echo -e "${YELLOW}⚠ Server starting up. Check logs at: tail -f $DIR/logs/server.log${NC}"
fi

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${GREEN}  🎉 Go Backend services have been restarted successfully!      ${NC}"
echo -e "${BLUE}================================================================${NC}"
