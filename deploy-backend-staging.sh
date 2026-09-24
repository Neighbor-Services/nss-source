#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) - Staging Backend Deployment Script
# Direct On-Server Go Compilation & Systemd Service (No Docker required)
# Usage: sudo bash deploy-backend-staging.sh
# ==============================================================================

set -e

# Configuration
APP_NAME="ns-backend-staging"
APP_DIR="/opt/ns-staging/backend"
USER="afari"
GROUP="www-data"
BACKEND_PORT=8001
DEFAULT_DOMAIN="staging-api.neighborservice.com"

# Database Configuration Defaults
DB_NAME=""
DB_USER=""
DB_PASSWORD=""

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
    echo -e "${RED}Please run as root (use sudo bash deploy-backend-staging.sh)${NC}"
    exit 1
fi

echo -e "${BLUE}=====================================================${NC}"
echo -e "${CYAN}  Neighbor Service (NSA) - Staging Backend Deploy     ${NC}"
echo -e "${BLUE}=====================================================${NC}"
echo ""

echo -e "${YELLOW}Phase 1: System Dependencies & Toolchain${NC}"

# Detect running user if afari doesn't exist
if ! id "$USER" &>/dev/null; then
    if [ -n "$SUDO_USER" ]; then
        USER="$SUDO_USER"
    else
        USER="nsapp"
        echo -e "${YELLOW}Creating service user $USER...${NC}"
        useradd -m -s /bin/bash "$USER" || true
    fi
fi

# Ensure Go toolchain is installed
if ! command -v go &> /dev/null && [ ! -f "/usr/local/go/bin/go" ]; then
    echo -e "${YELLOW}Go is not installed. Installing Go 1.24...${NC}"
    wget -q https://go.dev/dl/go1.24.0.linux-amd64.tar.gz || wget -q https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go*.linux-amd64.tar.gz
    rm -f go*.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo "export PATH=\$PATH:/usr/local/go/bin" >> /etc/profile
    echo -e "${GREEN}✓ Go installed successfully.${NC}"
else
    export PATH=$PATH:/usr/local/go/bin
    echo -e "${GREEN}✓ Go toolchain detected: $(go version)${NC}"
fi

# Ensure system packages
echo -e "${YELLOW}Ensuring system packages (git, curl, make, postgresql, nginx, ufw, rsync)...${NC}"
apt-get update -qq
apt-get install -y -qq git curl make postgresql postgresql-contrib nginx rsync

echo -e "${YELLOW}Phase 2: Directory & Code Placement${NC}"

mkdir -p /opt/ns
mkdir -p "$APP_DIR/bin"
mkdir -p "$APP_DIR/media"
mkdir -p "$APP_DIR/logs"
chmod 755 /opt /opt/ns "$APP_DIR"
chown -R $USER:$GROUP "$APP_DIR"

echo -e "${YELLOW}Copying Go backend files to $APP_DIR...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -d "$SCRIPT_DIR/backend-go" ] && [ -f "$SCRIPT_DIR/backend-go/cmd/server/main.go" ]; then
    SRC_DIR="$SCRIPT_DIR/backend-go"
elif [ -d "/home/afari/Projects/ns/backend-go" ] && [ -f "/home/afari/Projects/ns/backend-go/cmd/server/main.go" ]; then
    SRC_DIR="/home/afari/Projects/ns/backend-go"
elif [ -f "$SCRIPT_DIR/cmd/server/main.go" ]; then
    SRC_DIR="$SCRIPT_DIR"
else
    SRC_DIR=$(find /home /root /opt /var/www -maxdepth 4 -name "backend-go" -type d 2>/dev/null | while read -r d; do [ -f "$d/cmd/server/main.go" ] && echo "$d" && break; done)
fi

if [ -z "$SRC_DIR" ] || [ ! -f "$SRC_DIR/cmd/server/main.go" ]; then
    echo -e "${RED}Error: backend-go source directory (containing cmd/server/main.go) not found!${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Found source at: $SRC_DIR${NC}"
rsync -a --delete --exclude='bin' --exclude='.git' "$SRC_DIR/" "$APP_DIR/"
chown -R $USER:$GROUP "$APP_DIR"
chmod -R u+rwX,g+rX,o+rX "$APP_DIR"

cd "$APP_DIR"

echo -e "${YELLOW}Phase 2.5: PostgreSQL Staging Database Setup${NC}"

echo -e "${YELLOW}Do you want to configure the Staging PostgreSQL database?${NC}"
echo "   1) Yes - Create database, user and grant permissions"
echo "   2) No  - Skip database setup"
read -p "   Enter choice [1-2, default 2]: " DB_SETUP_CHOICE
DB_SETUP_CHOICE=${DB_SETUP_CHOICE:-2}

if [ "$DB_SETUP_CHOICE" = "1" ]; then
    read -p "Enter Staging Database Name [default: $DB_NAME]: " INPUT_DB_NAME
    DB_NAME=${INPUT_DB_NAME:-$DB_NAME}
    
    read -p "Enter Database User [default: $DB_USER]: " INPUT_DB_USER
    DB_USER=${INPUT_DB_USER:-$DB_USER}
    
    if [ -z "$DB_PASSWORD" ]; then
        read -s -p "Enter Database Password: " DB_PASSWORD
        echo ""
    fi
    
    echo -e "${YELLOW}Configuring PostgreSQL database and user...${NC}"
    systemctl start postgresql
    systemctl enable postgresql
    
    sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
        sudo -u postgres psql -c "CREATE DATABASE $DB_NAME;"
        
    sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE usename = '$DB_USER'" | grep -q 1 || \
        sudo -u postgres psql -c "CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';"
        
    sudo -u postgres psql -c "ALTER USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';"
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
    sudo -u postgres psql -c "ALTER DATABASE $DB_NAME OWNER TO $DB_USER;"
    echo -e "${GREEN}✓ PostgreSQL staging database '$DB_NAME' configured successfully${NC}"
fi

echo -e "${YELLOW}Phase 3: Environment Configuration${NC}"

if [ ! -f "$APP_DIR/.env" ]; then
    if [ -f "$APP_DIR/.env.example" ]; then
        echo -e "${YELLOW}Creating .env from .env.example...${NC}"
        cp "$APP_DIR/.env.example" "$APP_DIR/.env"
    elif [ -f "$SRC_DIR/.env" ]; then
        echo -e "${YELLOW}Copying existing .env...${NC}"
        cp "$SRC_DIR/.env" "$APP_DIR/.env"
    elif [ -f "$SCRIPT_DIR/backend-go/.env" ]; then
        echo -e "${YELLOW}Copying existing .env...${NC}"
        cp "$SCRIPT_DIR/backend-go/.env" "$APP_DIR/.env"
    fi
    echo -e "${CYAN}Please verify database credentials & secrets in $APP_DIR/.env${NC}"
else
    echo -e "${GREEN}✓ Existing .env preserved at $APP_DIR/.env${NC}"
fi

# Always enforce staging PORT (8001) and environment in .env
if grep -q '^PORT=' "$APP_DIR/.env"; then
    sed -i "s/^PORT=.*/PORT=$BACKEND_PORT/" "$APP_DIR/.env"
else
    echo "PORT=$BACKEND_PORT" >> "$APP_DIR/.env"
fi

if grep -q '^ENV=' "$APP_DIR/.env"; then
    sed -i 's/^ENV=.*/ENV=staging/' "$APP_DIR/.env"
else
    echo "ENV=staging" >> "$APP_DIR/.env"
fi

if grep -q '^ENVIRONMENT=' "$APP_DIR/.env"; then
    sed -i 's/^ENVIRONMENT=.*/ENVIRONMENT=staging/' "$APP_DIR/.env"
else
    echo "ENVIRONMENT=staging" >> "$APP_DIR/.env"
fi

chown $USER:$GROUP "$APP_DIR/.env"
chmod 600 "$APP_DIR/.env"

echo -e "${YELLOW}Phase 4: Direct Compilation (Go Binary Build)${NC}"

export PATH=$PATH:/usr/local/go/bin

echo -e "${YELLOW}Downloading Go dependencies...${NC}"
sudo -u $USER bash -c "cd '$APP_DIR' && env PATH=\"\$PATH:/usr/local/go/bin\" go mod tidy"

echo -e "${YELLOW}Compiling Neighbor Service Staging API Server (cmd/server/main.go)...${NC}"
sudo -u $USER bash -c "cd '$APP_DIR' && env PATH=\"\$PATH:/usr/local/go/bin\" go build -ldflags='-s -w' -o bin/server ./cmd/server/"
chmod +x "$APP_DIR/bin/server"
echo -e "${GREEN}✓ Staging API server compiled: $APP_DIR/bin/server${NC}"

echo -e "${YELLOW}Compiling Admin Generator CLI (cmd/create_admin/main.go)...${NC}"
sudo -u $USER bash -c "cd '$APP_DIR' && env PATH=\"\$PATH:/usr/local/go/bin\" go build -ldflags='-s -w' -o bin/create_admin ./cmd/create_admin/"
chmod +x "$APP_DIR/bin/create_admin"
echo -e "${GREEN}✓ Admin generator CLI compiled: $APP_DIR/bin/create_admin${NC}"

# Superuser / Administrator Account Creation Prompt
echo ""
echo -e "${YELLOW}Do you want to create or update a Staging Admin Account now?${NC}"
echo "   1) Yes - Run Admin Account Creator"
echo "   2) No  - Skip admin creation"
read -p "   Enter choice [1-2, default 2]: " ADMIN_CHOICE
ADMIN_CHOICE=${ADMIN_CHOICE:-2}

if [ "$ADMIN_CHOICE" = "1" ]; then
    read -p "Enter Admin Email [default: staging-admin@neighborservice.com]: " ADMIN_EMAIL
    ADMIN_EMAIL=${ADMIN_EMAIL:-staging-admin@neighborservice.com}
    read -s -p "Enter Admin Password [default: Admin123!]: " ADMIN_PASS
    ADMIN_PASS=${ADMIN_PASS:-Admin123!}
    echo ""
    read -p "Enter First Name [default: Staging]: " ADMIN_FN
    ADMIN_FN=${ADMIN_FN:-Staging}
    read -p "Enter Last Name [default: Admin]: " ADMIN_LN
    ADMIN_LN=${ADMIN_LN:-Admin}
    
    sudo -u $USER env PATH="$PATH" "$APP_DIR/bin/create_admin" \
      -email="$ADMIN_EMAIL" \
      -password="$ADMIN_PASS" \
      -first-name="$ADMIN_FN" \
      -last-name="$ADMIN_LN" || echo -e "${YELLOW}Admin creation step completed.${NC}"
fi

echo -e "${YELLOW}Phase 5: Systemd Service Setup${NC}"

cat << EOF > /etc/systemd/system/$APP_NAME.service
[Unit]
Description=Neighbor Service (NSA) Staging Backend API Server
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/bin/server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
LimitNOFILE=65535
EnvironmentFile=$APP_DIR/.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable $APP_NAME
systemctl restart $APP_NAME
echo -e "${GREEN}✓ Systemd service $APP_NAME enabled and active on port $BACKEND_PORT${NC}"

echo -e "${YELLOW}Phase 6: Nginx Gateway & SSL Configuration${NC}"

read -p "Do you want to configure Nginx reverse proxy with SSL for the Staging API? (y/n) [default: y]: " SETUP_SSL
SETUP_SSL=${SETUP_SSL:-y}

if [ "$SETUP_SSL" = "y" ]; then
    read -p "Enter your Staging API Domain (e.g. $DEFAULT_DOMAIN): " API_DOMAIN
    API_DOMAIN=${API_DOMAIN:-$DEFAULT_DOMAIN}
    
    mkdir -p "/etc/nginx/ssl/$API_DOMAIN"
    
    if [ ! -f "/etc/nginx/ssl/$API_DOMAIN/cert.pem" ]; then
        echo "Please paste your SSL Certificate (Cloudflare Origin Cert or Let's Encrypt fullchain.pem), then press Ctrl+D:"
        cat > "/etc/nginx/ssl/$API_DOMAIN/cert.pem"
    else
        echo -e "${GREEN}✓ SSL Certificate found at /etc/nginx/ssl/$API_DOMAIN/cert.pem${NC}"
    fi
    
    if [ ! -f "/etc/nginx/ssl/$API_DOMAIN/key.pem" ]; then
        echo "Please paste your SSL Private Key (key.pem or privkey.pem), then press Ctrl+D:"
        cat > "/etc/nginx/ssl/$API_DOMAIN/key.pem"
    else
        echo -e "${GREEN}✓ SSL Private Key found at /etc/nginx/ssl/$API_DOMAIN/key.pem${NC}"
    fi
    
    chmod 600 "/etc/nginx/ssl/$API_DOMAIN/key.pem"
    
    cat << EOF > /etc/nginx/sites-available/$APP_NAME
upstream ns_staging_cluster {
    server 127.0.0.1:$BACKEND_PORT max_fails=3 fail_timeout=10s;
    keepalive 32;
}

# Rate limiting zones
limit_req_zone \$binary_remote_addr zone=ns_stage_api_limit:10m rate=40r/s;
limit_req_zone \$binary_remote_addr zone=ns_stage_auth_limit:10m rate=10r/s;

server {
    listen 80;
    server_name $API_DOMAIN;
    server_tokens off;
    
    # Drop direct IP scans
    if (\$host ~* "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$") {
        return 444;
    }
    
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl;
    server_name $API_DOMAIN;
    server_tokens off;

    # Drop direct IP scans
    if (\$host ~* "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$") {
        return 444;
    }

    ssl_certificate /etc/nginx/ssl/$API_DOMAIN/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/$API_DOMAIN/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';

    # Security Headers
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

    # Media & Document Upload Limit
    client_max_body_size 50M;

    # Rate-limited Auth Endpoints
    location /api/v1/accounts/login/ {
        limit_req zone=ns_stage_auth_limit burst=15 nodelay;
        proxy_pass http://ns_staging_cluster;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
    }

    # Real-Time WebSocket Endpoint
    location /ws {
        proxy_pass http://ns_staging_cluster/ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }

    # API Proxy
    location / {
        limit_req zone=ns_stage_api_limit burst=60 nodelay;
        proxy_pass http://ns_staging_cluster;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;

        proxy_connect_timeout 10s;
        proxy_send_timeout 120s;
        proxy_read_timeout 120s;
    }
}
EOF
else
    cat << EOF > /etc/nginx/sites-available/$APP_NAME
server {
    listen 80;
    server_name _;
    server_tokens off;

    client_max_body_size 50M;

    location / {
        proxy_pass http://127.0.0.1:$BACKEND_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF
fi

ln -sf /etc/nginx/sites-available/$APP_NAME /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx || systemctl restart nginx
echo -e "${GREEN}✓ Staging Nginx API gateway configured and active${NC}"

# Optional Cloudflare Cache Purge
if [ -n "$CF_ZONE_ID" ] && [ -n "$CF_API_TOKEN" ]; then
    echo -e "${YELLOW}Purging Cloudflare cache for Staging API...${NC}"
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
echo -e "${GREEN}✓ Staging Backend Deployment Completed Successfully!${NC}"
echo -e "${BLUE}=====================================================${NC}"
echo -e "${CYAN}Staging API:   ${NC} https://$DEFAULT_DOMAIN"
echo -e "${CYAN}Service Status:${NC} sudo systemctl status $APP_NAME"
echo -e "${CYAN}Live Logs:     ${NC} sudo journalctl -u $APP_NAME -f"
echo ""
