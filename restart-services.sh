#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) — Universal Services Restart Script
# Centralized script to restart all backend services, replicas, workers,
# staging environments, Redis, and Nginx.
#
# Usage:
#   sudo bash restart-services.sh          (Restart everything)
#   sudo bash restart-services.sh --go     (Restart only Go cluster)
#   sudo bash restart-services.sh --django (Restart only Django & Celery)
#   sudo bash restart-services.sh --redis  (Restart only Redis instances)
#   sudo bash restart-services.sh --nginx  (Restart only Nginx)
# ==============================================================================

set -e

# Colors for terminal output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}   Neighbor Service (NSA) — Universal Service Manager           ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

# Check if running as root for system services
IS_ROOT=false
if [ "$EUID" -eq 0 ]; then
    IS_ROOT=true
fi

MODE="${1:-all}"

restart_redis() {
    echo -e "${YELLOW}▶ Restarting Redis Instances...${NC}"
    if [ "$IS_ROOT" = true ]; then
        if systemctl is-active --quiet redis-server || systemctl is-enabled --quiet redis-server 2>/dev/null; then
            systemctl restart redis-server && echo -e "${GREEN}  ✓ Primary redis-server restarted${NC}" || echo -e "${RED}  ✗ Failed to restart primary redis${NC}"
        fi
        if systemctl is-active --quiet redis || systemctl is-enabled --quiet redis 2>/dev/null; then
            systemctl restart redis && echo -e "${GREEN}  ✓ redis service restarted${NC}" || true
        fi

        # Secondary Redis Multi-instances (redis2, redis3, redis6380, etc.)
        for svc in $(systemctl list-units --type=service --all 2>/dev/null | grep -oE "redis[0-9_a-zA-Z-]+" | sort -u); do
            if [ "$svc" != "redis-server" ] && [ "$svc" != "redis" ]; then
                echo -e "  Restarting $svc..."
                systemctl restart "$svc" 2>/dev/null && echo -e "${GREEN}  ✓ $svc restarted${NC}" || true
            fi
        done
    else
        echo -e "${YELLOW}  (Running as non-root; skipping systemctl redis restart)${NC}"
    fi
}

restart_go_cluster() {
    echo -e "${YELLOW}▶ Restarting Go Backend Multi-Instance Cluster...${NC}"
    if [ "$IS_ROOT" = true ]; then
        # Check systemd target
        if systemctl list-unit-files | grep -q "ns-backend.target"; then
            systemctl restart ns-backend.target && echo -e "${GREEN}  ✓ ns-backend.target cluster restarted${NC}" || true
        fi

        # Restart individual active instances (ns-backend@1, ns-backend@2, etc.)
        for inst in $(systemctl list-units --type=service --all 2>/dev/null | grep -oE "ns-backend@[0-9]+" | sort -u); do
            echo -e "  Restarting instance $inst..."
            systemctl restart "$inst" && echo -e "${GREEN}  ✓ $inst restarted${NC}" || true
        done

        # Staging Go backend
        if systemctl list-unit-files | grep -q "ns-backend-staging"; then
            systemctl restart ns-backend-staging && echo -e "${GREEN}  ✓ ns-backend-staging restarted${NC}" || true
        fi
    else
        # Local development restart helper
        echo -e "${CYAN}  Checking local running Go server processes...${NC}"
        pkill -f "go run ./cmd/server/main.go" 2>/dev/null && echo -e "${GREEN}  ✓ Stopped local go run process${NC}" || true
        pkill -f "./bin/server" 2>/dev/null && echo -e "${GREEN}  ✓ Stopped compiled Go binary${NC}" || true
    fi
}

restart_django_celery() {
    echo -e "${YELLOW}▶ Restarting Django & Celery Workers (Supervisor)...${NC}"
    if command -v supervisorctl &>/dev/null; then
        # Production replicas
        supervisorctl restart ns_backend_replicas:* 2>/dev/null && echo -e "${GREEN}  ✓ ns_backend_replicas restarted${NC}" || true
        supervisorctl restart ns_backend:* 2>/dev/null && echo -e "${GREEN}  ✓ ns_backend group restarted${NC}" || true
        
        # Celery Workers & Beat
        supervisorctl restart ns_backend_celery 2>/dev/null && echo -e "${GREEN}  ✓ ns_backend_celery restarted${NC}" || true
        supervisorctl restart ns_backend_celery_beat 2>/dev/null && echo -e "${GREEN}  ✓ ns_backend_celery_beat restarted${NC}" || true
        
        # Staging Supervisor services
        supervisorctl restart ns_backend_staging:* 2>/dev/null && echo -e "${GREEN}  ✓ ns_backend_staging restarted${NC}" || true
        supervisorctl restart ns_staging:* 2>/dev/null && echo -e "${GREEN}  ✓ ns_staging restarted${NC}" || true
    else
        echo -e "${YELLOW}  Supervisor not detected; skipping supervisorctl${NC}"
    fi
}

restart_nginx() {
    echo -e "${YELLOW}▶ Verifying and Reloading Nginx...${NC}"
    if [ "$IS_ROOT" = true ] && command -v nginx &>/dev/null; then
        if nginx -t 2>/dev/null; then
            systemctl reload nginx && echo -e "${GREEN}  ✓ Nginx configuration tested and reloaded successfully${NC}" || systemctl restart nginx
        else
            echo -e "${RED}  ✗ Nginx syntax test failed; skipping reload to protect uptime${NC}"
        fi
    fi
}

# Execution Router
case "$MODE" in
    --go)
        restart_go_cluster
        ;;
    --django)
        restart_django_celery
        ;;
    --redis)
        restart_redis
        ;;
    --nginx)
        restart_nginx
        ;;
    *)
        restart_redis
        restart_go_cluster
        restart_django_celery
        restart_nginx
        ;;
esac

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${GREEN}  🎉 All specified services have been refreshed successfully!   ${NC}"
echo -e "${BLUE}================================================================${NC}"
