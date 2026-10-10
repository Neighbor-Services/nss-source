#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) - Master Staging Deployment Orchestrator
# Executes full-stack staging deployment in a single run:
#   1. Staging Backend Go API Server (Port 8001 / Staging DB)
#   2. Staging Public Website SPA (Angular Build & Nginx Hosting)
#   3. Staging Admin Operations Portal SPA (Angular Build & Nginx Hosting)
#
# Usage: sudo bash deploy-all-staging.sh
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use: sudo bash deploy-all-staging.sh)${NC}"
    exit 1
fi

echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}    NEIGHBOR SERVICE (NSA) - FULL-STACK STAGING DEPLOYMENT      ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""
echo -e "${YELLOW}This master orchestrator will deploy all 3 staging components:${NC}"
echo -e "   1. ${BOLD}Staging Backend API${NC}     (https://staging-api.neighborservice.com on port 8001)"
echo -e "   2. ${BOLD}Staging Public Website${NC}  (https://staging.neighborservice.com)"
echo -e "   3. ${BOLD}Staging Admin Portal${NC}    (https://staging-admin.neighborservice.com)"
echo ""
read -p "Do you want to proceed with full-stack staging deployment? (y/n) [default: y]: " CONFIRM_DEPLOY
CONFIRM_DEPLOY=${CONFIRM_DEPLOY:-y}

if [ "$CONFIRM_DEPLOY" != "y" ] && [ "$CONFIRM_DEPLOY" != "Y" ]; then
    echo -e "${YELLOW}Staging deployment cancelled by user.${NC}"
    exit 0
fi

START_TIME=$(date +%s)

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}>>> STEP 1 / 3: Deploying Staging Backend API...${NC}"
echo -e "${BLUE}================================================================${NC}"
bash "$SCRIPT_DIR/deploy-backend-staging.sh"

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}>>> STEP 2 / 3: Deploying Staging Public Website...${NC}"
echo -e "${BLUE}================================================================${NC}"
bash "$SCRIPT_DIR/deploy-frontend-staging.sh"

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}>>> STEP 3 / 3: Deploying Staging Admin Portal...${NC}"
echo -e "${BLUE}================================================================${NC}"
bash "$SCRIPT_DIR/deploy-admin-frontend-staging.sh"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Final Nginx test & reload
echo -e "${YELLOW}Performing final Nginx validation...${NC}"
nginx -t && systemctl reload nginx

echo ""
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}${BOLD}    ✨ FULL-STACK STAGING DEPLOYMENT COMPLETED SUCCESSFULLY!    ${NC}"
echo -e "${GREEN}================================================================${NC}"
echo -e "${CYAN}Total Deployment Time:${NC} ${DURATION}s"
echo ""
echo -e "${BOLD}Staging Service Summary:${NC}"
echo -e "  🌐 ${CYAN}Staging Website:${NC}     https://staging.neighborservice.com"
echo -e "  🛡️ ${CYAN}Staging Admin:${NC}       https://staging-admin.neighborservice.com"
echo -e "  ⚡ ${CYAN}Staging API:${NC}         https://staging-api.neighborservice.com"
echo ""
echo -e "${BOLD}Management Commands:${NC}"
echo -e "  • Staging Service:    ${CYAN}sudo systemctl status ns-backend-staging${NC}"
echo -e "  • Staging Restart:    ${CYAN}sudo systemctl restart ns-backend-staging${NC}"
echo -e "  • Staging Logs:       ${CYAN}sudo journalctl -u ns-backend-staging -f${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""
