#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) - Master Production Deployment Orchestrator
# Executes full-stack production deployment in a single run:
#   1. Backend Go Cluster (Multi-Instance Systemd & Nginx Load Balancer)
#   2. Public Website SPA (Angular Build & Nginx Hosting)
#   3. Admin Operations Portal SPA (Angular Build & Nginx Hosting)
#
# Usage: sudo bash deploy-all.sh
# ==============================================================================

set -e

# Trap errors so failures in sub-scripts are clearly reported
trap 'echo -e "\033[0;31m\n[ERROR] deploy-all.sh aborted at line $LINENO (exit code $?). Check output above for details.\033[0m" >&2' ERR

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
    echo -e "${RED}Please run as root (use: sudo bash deploy-all.sh)${NC}"
    exit 1
fi

echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}   NEIGHBOR SERVICE (NSA) - FULL-STACK PRODUCTION DEPLOYMENT   ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""
echo -e "${YELLOW}This master orchestrator will deploy all 3 production components:${NC}"
echo -e "   1. ${BOLD}Backend Go Cluster${NC}      (Multi-Instance API on https://api.neighborservice.com)"
echo -e "   2. ${BOLD}Public Website SPA${NC}      (Landing & Web App on https://neighborservice.com)"
echo -e "   3. ${BOLD}Admin Operations Portal${NC} (Command Center on https://admin.neighborservice.com)"
echo ""
read -p "Do you want to proceed with full-stack production deployment? (y/n) [default: y]: " CONFIRM_DEPLOY
CONFIRM_DEPLOY=${CONFIRM_DEPLOY:-y}

if [ "$CONFIRM_DEPLOY" != "y" ] && [ "$CONFIRM_DEPLOY" != "Y" ]; then
    echo -e "${YELLOW}Deployment cancelled by user.${NC}"
    exit 0
fi

START_TIME=$(date +%s)

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}>>> STEP 1 / 3: Deploying Production Backend Cluster...${NC}"
echo -e "${BLUE}================================================================${NC}"
bash "$SCRIPT_DIR/deploy-backend.sh"

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}>>> STEP 2 / 3: Deploying Production Public Website...${NC}"
echo -e "${BLUE}================================================================${NC}"
bash "$SCRIPT_DIR/deploy-frontend.sh"

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}${BOLD}>>> STEP 3 / 3: Deploying Production Admin Portal...${NC}"
echo -e "${BLUE}================================================================${NC}"
bash "$SCRIPT_DIR/deploy-admin-frontend.sh"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Final Nginx test & reload
echo -e "${YELLOW}Performing final Nginx validation...${NC}"
nginx -t && systemctl reload nginx

echo ""
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}${BOLD}   ✨ FULL-STACK PRODUCTION DEPLOYMENT COMPLETED SUCCESSFULLY!  ${NC}"
echo -e "${GREEN}================================================================${NC}"
echo -e "${CYAN}Total Deployment Time:${NC} ${DURATION}s"
echo ""
echo -e "${BOLD}Production Service Summary:${NC}"
echo -e "  🌐 ${CYAN}Public Website:${NC}      https://neighborservice.com"
echo -e "  🛡️ ${CYAN}Admin Portal:${NC}        https://admin.neighborservice.com"
echo -e "  ⚡ ${CYAN}API Gateway:${NC}         https://api.neighborservice.com"
echo ""
echo -e "${BOLD}Management Commands:${NC}"
echo -e "  • Cluster Status:     ${CYAN}sudo systemctl status ns-backend.target${NC}"
echo -e "  • Cluster Restart:    ${CYAN}sudo systemctl restart ns-backend.target${NC}"
echo -e "  • Real-time Logs:     ${CYAN}sudo journalctl -u 'ns-backend@*' -f${NC}"
echo -e "  • Nginx Status:       ${CYAN}sudo systemctl status nginx${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""
