#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) — Deployment Teardown & Cleanup Script
# Safely stops and completely removes systemd services, supervisor replicas,
# Nginx site configurations, socket files, and logrotate configs created
# by the backend deployment scripts.
#
# Usage: sudo bash teardown-backend.sh
# ==============================================================================

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root: sudo bash teardown-backend.sh${NC}"
    exit 1
fi

echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}   Neighbor Service (NSA) — Backend Deployment Teardown         ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

echo -e "${YELLOW}1. Stopping and Disabling Go Backend Systemd Services...${NC}"

# Stop and disable systemd targets
if systemctl list-unit-files | grep -q "ns-backend.target"; then
    systemctl stop ns-backend.target 2>/dev/null || true
    systemctl disable ns-backend.target 2>/dev/null || true
    rm -f /etc/systemd/system/ns-backend.target
    echo -e "${GREEN}  ✓ Removed /etc/systemd/system/ns-backend.target${NC}"
fi

# Stop and disable individual service instances (ns-backend@1..20)
for inst in $(systemctl list-units --type=service --all 2>/dev/null | grep -oE "ns-backend@[0-9]+" | sort -u); do
    echo -e "  Stopping $inst..."
    systemctl stop "$inst" 2>/dev/null || true
    systemctl disable "$inst" 2>/dev/null || true
done

# Remove service template
if [ -f "/etc/systemd/system/ns-backend@.service" ]; then
    rm -f /etc/systemd/system/ns-backend@.service
    echo -e "${GREEN}  ✓ Removed /etc/systemd/system/ns-backend@.service${NC}"
fi

# Staging service
if [ -f "/etc/systemd/system/ns-backend-staging.service" ]; then
    systemctl stop ns-backend-staging 2>/dev/null || true
    systemctl disable ns-backend-staging 2>/dev/null || true
    rm -f /etc/systemd/system/ns-backend-staging.service
    echo -e "${GREEN}  ✓ Removed /etc/systemd/system/ns-backend-staging.service${NC}"
fi

systemctl daemon-reload
echo -e "${GREEN}✓ Systemd units cleaned up${NC}"

echo -e "${YELLOW}2. Stopping and Cleaning Supervisor Instances (Django & Celery)...${NC}"
if command -v supervisorctl &>/dev/null; then
    supervisorctl stop ns_backend_replicas:* 2>/dev/null || true
    supervisorctl stop ns_backend:* 2>/dev/null || true
    supervisorctl stop ns_backend_celery 2>/dev/null || true
    supervisorctl stop ns_backend_celery_beat 2>/dev/null || true
    supervisorctl stop ns_backend_staging:* 2>/dev/null || true

    rm -f /etc/supervisor/conf.d/ns_backend*.conf
    rm -f /etc/supervisor/conf.d/ns-backend*.conf
    supervisorctl reread 2>/dev/null || true
    supervisorctl update 2>/dev/null || true
    echo -e "${GREEN}✓ Supervisor configurations removed${NC}"
fi

echo -e "${YELLOW}3. Removing Nginx Site Configurations & Sockets...${NC}"
rm -f /etc/nginx/sites-enabled/ns-backend*
rm -f /etc/nginx/sites-available/ns-backend*
rm -f /etc/nginx/sites-enabled/ns_backend*
rm -f /etc/nginx/sites-available/ns_backend*
rm -f /run/ns_backend*.sock
rm -f /tmp/ns_backend*.sock

if command -v nginx &>/dev/null; then
    nginx -t 2>/dev/null && systemctl reload nginx || systemctl restart nginx
    echo -e "${GREEN}✓ Nginx sites removed and Nginx reloaded${NC}"
fi

echo -e "${YELLOW}4. Removing Logrotate Configurations...${NC}"
rm -f /etc/logrotate.d/ns-backend*
rm -f /etc/logrotate.d/ns_backend*
echo -e "${GREEN}✓ Logrotate configs removed${NC}"

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${GREEN}  ✓ All deployment artifacts and services successfully removed! ${NC}"
echo -e "${BLUE}================================================================${NC}"
