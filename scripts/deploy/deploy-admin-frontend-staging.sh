#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) - Staging Admin Portal Deployment Script
# Direct Angular SPA Build & Nginx Static Hosting (No Docker required)
# Usage: sudo bash deploy-admin-frontend-staging.sh
# ==============================================================================

set -e

# Configuration
APP_NAME="ns-admin-staging"
BUILD_DIR="/opt/ns-staging/admin-build"
WEB_DIR="/var/www/ns-admin-staging"
USER="afari"
GROUP="www-data"
DEFAULT_DOMAIN="staging-admin.neighborservice.com"
API_BASE_URL="https://staging-api.neighborservice.com/api/v1"

# Cloudflare Cache Purge (optional)
CF_ZONE_ID=""
CF_API_TOKEN=""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo bash deploy-admin-frontend-staging.sh)${NC}"
    exit 1
fi

echo -e "${BLUE}=====================================================${NC}"
echo -e "${CYAN}  Neighbor Service (NSA) - Staging Admin Deploy      ${NC}"
echo -e "${BLUE}=====================================================${NC}"
echo ""

echo -e "${YELLOW}Phase 1: System Dependencies & Node.js Environment${NC}"

# Detect running user
if ! id "$USER" &>/dev/null; then
    if [ -n "$SUDO_USER" ]; then
        USER="$SUDO_USER"
    else
        USER="nsapp"
        echo -e "${YELLOW}Creating service user $USER...${NC}"
        useradd -m -s /bin/bash "$USER" || true
    fi
fi

# Ensure Node.js & npm are installed
if ! command -v node &> /dev/null; then
    echo -e "${YELLOW}Node.js is not installed. Installing Node.js 22 LTS...${NC}"
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
    apt-get install -y -qq nodejs
    echo -e "${GREEN}✓ Node.js installed successfully.${NC}"
fi

export PATH=$PATH:/usr/bin:/usr/local/bin:/snap/bin

# Check nvm locations if applicable
SUDO_USER_HOME=$(getent passwd "${SUDO_USER:-$USER}" | cut -d: -f6)
for NVM_DIR in "$SUDO_USER_HOME/.nvm" "/root/.nvm"; do
    if [ -d "$NVM_DIR" ]; then
        NVM_NODE=$(ls "$NVM_DIR/versions/node" 2>/dev/null | sort -V | tail -1)
        [ -n "$NVM_NODE" ] && export PATH="$NVM_DIR/versions/node/$NVM_NODE/bin:$PATH"
        break
    fi
done

if ! command -v npm &> /dev/null; then
    echo -e "${RED}npm not found. Please install Node.js system-wide:${NC}"
    echo "  curl -fsSL https://deb.nodesource.com/setup_22.x | bash -"
    echo "  apt-get install -y nodejs"
    exit 1
fi

echo -e "${GREEN}✓ Node $(node -v) / npm $(npm -v) active${NC}"

# Ensure Nginx & rsync are installed
apt-get update -qq
apt-get install -y -qq nginx rsync

echo -e "${YELLOW}Phase 2: Code Placement & Build Directory${NC}"

mkdir -p "$BUILD_DIR"
chown -R $USER:$GROUP "$BUILD_DIR"

echo -e "${YELLOW}Copying Admin source files to $BUILD_DIR...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -d "$SCRIPT_DIR/frontend/admin-portal" ]; then
    rsync -a --delete --exclude='node_modules' --exclude='.angular' --exclude='dist' "$SCRIPT_DIR/frontend/admin-portal/" "$BUILD_DIR/"
elif [ -d "$SCRIPT_DIR/admin-portal" ]; then
    rsync -a --delete --exclude='node_modules' --exclude='.angular' --exclude='dist' "$SCRIPT_DIR/admin-portal/" "$BUILD_DIR/"
else
    rsync -a --delete --exclude='node_modules' --exclude='.angular' --exclude='dist' ./ "$BUILD_DIR/"
fi
chown -R $USER:$GROUP "$BUILD_DIR"

cd "$BUILD_DIR"

echo -e "${YELLOW}Phase 3: Building Staging Admin Angular SPA${NC}"

RESOLVED_PATH="$PATH"
if [ -d "/home/$USER/.npm" ]; then
    chown -R $USER:$USER "/home/$USER/.npm" 2>/dev/null || true
fi

# Ensure API endpoint is configured for Staging API Subdomain
read -p "Enter Staging Backend API Base URL [default: $API_BASE_URL]: " INPUT_API_URL
API_BASE_URL=${INPUT_API_URL:-$API_BASE_URL}

echo -e "${YELLOW}Configuring Staging Admin API target: $API_BASE_URL${NC}"
mkdir -p "$BUILD_DIR/public"
cat << EOF > "$BUILD_DIR/public/env.js"
window.__API_BASE_URL__ = "$API_BASE_URL";
EOF

echo -e "${YELLOW}Installing npm packages...${NC}"
if [ -f "package-lock.json" ]; then
    sudo -u $USER env PATH="$RESOLVED_PATH" npm ci --legacy-peer-deps || sudo -u $USER env PATH="$RESOLVED_PATH" npm install --legacy-peer-deps
else
    sudo -u $USER env PATH="$RESOLVED_PATH" npm install --legacy-peer-deps
fi

echo -e "${YELLOW}Executing Admin Angular production build (ng build)...${NC}"
sudo -u $USER env PATH="$RESOLVED_PATH" npm run build

# Detect build output directory
DIST_DIR=""
if [ -d "$BUILD_DIR/dist/admin-portal/browser" ]; then
    DIST_DIR="$BUILD_DIR/dist/admin-portal/browser"
elif [ -d "$BUILD_DIR/dist/admin-portal" ]; then
    DIST_DIR="$BUILD_DIR/dist/admin-portal"
elif [ -d "$BUILD_DIR/dist/browser" ]; then
    DIST_DIR="$BUILD_DIR/dist/browser"
fi

if [ -z "$DIST_DIR" ]; then
    echo -e "${RED}Admin build output directory not found! Check build logs above.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Staging Admin production bundle generated at $DIST_DIR${NC}"

echo -e "${YELLOW}Phase 4: Deploying to Staging Admin Web Root ($WEB_DIR)${NC}"

mkdir -p "$WEB_DIR"
rm -rf "$WEB_DIR"/*
cp -a "$DIST_DIR/." "$WEB_DIR/"
# Ensure env.js exists in root
cat << EOF > "$WEB_DIR/env.js"
window.__API_BASE_URL__ = "$API_BASE_URL";
EOF

chown -R www-data:www-data "$WEB_DIR"
chmod -R 755 "$WEB_DIR"
echo -e "${GREEN}✓ Staging Admin static assets deployed to $WEB_DIR${NC}"

echo -e "${YELLOW}Phase 5: Nginx Web Server Configuration${NC}"

read -p "Do you want to configure Nginx with SSL for the Staging Admin Portal? (y/n) [default: y]: " SETUP_SSL
SETUP_SSL=${SETUP_SSL:-y}

if [ "$SETUP_SSL" = "y" ]; then
    read -p "Enter your staging admin domain (e.g. $DEFAULT_DOMAIN): " ADMIN_DOMAIN
    ADMIN_DOMAIN=${ADMIN_DOMAIN:-$DEFAULT_DOMAIN}
    
    # Also check parent root domain for wildcard SSL directory
    ROOT_DOMAIN=$(echo "$ADMIN_DOMAIN" | sed -e 's/^admin\.//' -e 's/^staging-admin\.//')
    SSL_DIR="/etc/nginx/ssl/$ROOT_DOMAIN"
    [ ! -d "$SSL_DIR" ] && SSL_DIR="/etc/nginx/ssl/$ADMIN_DOMAIN"
    mkdir -p "$SSL_DIR"
    
    if [ ! -f "$SSL_DIR/cert.pem" ]; then
        echo "Please paste your SSL Certificate (Cloudflare Origin Cert or Let's Encrypt fullchain.pem), then press Ctrl+D:"
        cat > "$SSL_DIR/cert.pem"
    else
        echo -e "${GREEN}✓ SSL Certificate found at $SSL_DIR/cert.pem${NC}"
    fi
    
    if [ ! -f "$SSL_DIR/key.pem" ]; then
        echo "Please paste your SSL Private Key (key.pem or privkey.pem), then press Ctrl+D:"
        cat > "$SSL_DIR/key.pem"
    else
        echo -e "${GREEN}✓ SSL Private Key found at $SSL_DIR/key.pem${NC}"
    fi
    
    chmod 600 "$SSL_DIR/key.pem"
    
    cat << EOF > /etc/nginx/sites-available/$APP_NAME
server {
    listen 80;
    server_name $ADMIN_DOMAIN;
    server_tokens off;
    
    # Drop direct IP scans
    if (\$host ~* "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$") {
        return 444;
    }
    
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl;
    server_name $ADMIN_DOMAIN;
    server_tokens off;

    root $WEB_DIR;
    index index.html;

    # Drop direct IP scans
    if (\$host ~* "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$") {
        return 444;
    }

    ssl_certificate $SSL_DIR/cert.pem;
    ssl_certificate_key $SSL_DIR/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';

    # Hardened Security Headers for Admin Portal
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/x-javascript application/json application/xml image/svg+xml;

    client_max_body_size 50M;

    # Static Assets Caching
    location ~* \.(?:css|js|jpg|jpeg|gif|png|ico|svg|woff|woff2|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public, no-transform";
        try_files \$uri =404;
    }

    # Admin SPA HTML5 Routing
    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
EOF
else
    cat << EOF > /etc/nginx/sites-available/$APP_NAME
server {
    listen 80;
    server_name _;

    root $WEB_DIR;
    index index.html;

    gzip on;
    gzip_types text/plain text/css application/javascript application/json image/svg+xml;

    client_max_body_size 50M;

    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
EOF
fi

ln -sf /etc/nginx/sites-available/$APP_NAME /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx || systemctl restart nginx
echo -e "${GREEN}✓ Staging Admin Nginx static hosting configured and active${NC}"

# Optional Cloudflare Cache Purge
if [ -n "$CF_ZONE_ID" ] && [ -n "$CF_API_TOKEN" ]; then
    echo -e "${YELLOW}Purging Cloudflare cache for Staging Admin...${NC}"
    CF_RESPONSE=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CF_ZONE_ID/purge_cache" \
        -H "Authorization: Bearer $CF_API_TOKEN" \
        -H "Content-Type: application/json" \
        --data '{"purge_everything":true}')
    if echo "$CF_RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Cloudflare cache purged successfully${NC}"
    else
        echo -e "${YELLOW}⚠ Cloudflare purge status: $CF_RESPONSE${NC}"
    fi
fi

echo ""
echo -e "${BLUE}=====================================================${NC}"
echo -e "${GREEN}✓ Staging Admin Deployment Completed Successfully!${NC}"
echo -e "${BLUE}=====================================================${NC}"
echo -e "${CYAN}Staging Admin:  ${NC} https://$DEFAULT_DOMAIN"
echo -e "${CYAN}API Subdomain:  ${NC} $API_BASE_URL"
echo -e "${CYAN}Admin Web Root: ${NC} $WEB_DIR"
echo -e "${CYAN}Nginx Site:     ${NC} /etc/nginx/sites-available/$APP_NAME"
echo ""
