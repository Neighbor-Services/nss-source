#!/bin/bash
# ==============================================================================
# Neighbor Service (NSA) - Production Backend Deployment Script
# Multi-Instance Go Cluster & Nginx Upstream Load Balancing (No Docker required)
# Usage: sudo bash deploy-backend.sh
# ==============================================================================

set -e

# Trap errors and print the line number for easier debugging
trap 'echo -e "\033[0;31m[ERROR] deploy-backend.sh failed at line $LINENO — exit code $?\033[0m" >&2' ERR

# Configuration
APP_NAME="ns-backend"
APP_DIR="/opt/ns/backend"
USER="afari"
GROUP="www-data"
DEFAULT_INSTANCES=5
BASE_PORT=8010 # Instances will run on 8011, 8012, 8013, ...
DEFAULT_DOMAIN="api.neighborservice.com"

# Database Configuration Defaults
DB_NAME="ns_db"
DB_USER="postgres"
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
    echo -e "${RED}Please run as root (use sudo bash deploy-backend.sh)${NC}"
    exit 1
fi

echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}  Neighbor Service (NSA) - Production Backend Cluster Deploy    ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

echo -e "${YELLOW}Phase 1: Cluster Sizing & System Toolchain${NC}"

# Ask for number of instances (5 to 10 recommended)
read -p "How many backend API worker instances do you want to run for load balancing? (e.g. 5 or 10) [default: $DEFAULT_INSTANCES]: " NUM_INSTANCES
NUM_INSTANCES=${NUM_INSTANCES:-$DEFAULT_INSTANCES}

if ! [[ "$NUM_INSTANCES" =~ ^[0-9]+$ ]] || [ "$NUM_INSTANCES" -lt 1 ]; then
    echo -e "${RED}Invalid instance count. Defaulting to $DEFAULT_INSTANCES.${NC}"
    NUM_INSTANCES=$DEFAULT_INSTANCES
fi

echo -e "${GREEN}✓ Cluster configuration: $NUM_INSTANCES backend instances (Ports $(($BASE_PORT + 1)) to $(($BASE_PORT + $NUM_INSTANCES)))${NC}"

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
    SRC_DIR=$(find /home /root /opt /var/www -maxdepth 4 -name "backend-go" -type d 2>/dev/null | while read -r d; do [ -f "$d/cmd/server/main.go" ] && echo "$d" && break; done || true)
    SRC_DIR=${SRC_DIR:-""}
fi

if [ -z "$SRC_DIR" ] || [ ! -f "$SRC_DIR/cmd/server/main.go" ]; then
    echo -e "${RED}Error: backend-go source directory not found.${NC}"
    echo -e "${YELLOW}Please manually place the source before running this script:${NC}"
    echo -e "  Option 1 — Copy from local machine:"
    echo -e "    ${CYAN}rsync -av /path/to/ns/backend-go/ root@<server-ip>:/opt/ns/backend-go/${NC}"
    echo -e "  Option 2 — Place alongside this script:"
    echo -e "    ${CYAN}The script expects:  $(dirname "$0")/backend-go/cmd/server/main.go${NC}"
    echo -e "  Option 3 — Clone with submodules on the server:"
    echo -e "    ${CYAN}git clone --recurse-submodules https://github.com/Neighbor-Services/nss-source.git /opt/ns/repo${NC}"
    echo -e "    ${CYAN}Then re-run this script from: /opt/ns/repo${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Found source at: $SRC_DIR${NC}"
# Preserve server-only files: .env, service account JSON credentials, media uploads, logs, and binaries
rsync -a \
    --exclude='bin' \
    --exclude='.git' \
    --exclude='.env*' \
    --exclude='media' \
    --exclude='logs' \
    --exclude='*service*account*.json' \
    --exclude='*serviceAccount*.json' \
    --exclude='firebase*.json' \
    --exclude='certs' \
    "$SRC_DIR/" "$APP_DIR/"
chown -R $USER:$GROUP "$APP_DIR" || true
chmod -R u+rwX,g+rX,o+rX "$APP_DIR" || true

cd "$APP_DIR"

echo -e "${YELLOW}Phase 2.5: PostgreSQL Database Setup${NC}"

echo -e "${YELLOW}Do you want to configure the PostgreSQL database?${NC}"
echo "   1) Yes - Create database, user and grant permissions"
echo "   2) No  - Skip database setup (already running or existing DB)"
read -p "   Enter choice [1-2, default 2]: " DB_SETUP_CHOICE
DB_SETUP_CHOICE=${DB_SETUP_CHOICE:-2}

if [ "$DB_SETUP_CHOICE" = "1" ]; then
    read -p "Enter Database Name [default: $DB_NAME]: " INPUT_DB_NAME
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
    echo -e "${GREEN}✓ PostgreSQL database '$DB_NAME' configured successfully${NC}"
fi

# Automated PostgreSQL Backup Configuration
echo -e "${YELLOW}Do you want to configure automated daily PostgreSQL backups?${NC}"
echo "   1) Yes - Configure daily backups (30-day retention at 2 AM)"
echo "   2) No  - Skip backup cron"
read -p "   Enter choice [1-2, default 2]: " BACKUP_CHOICE
BACKUP_CHOICE=${BACKUP_CHOICE:-2}

if [ "$BACKUP_CHOICE" = "1" ]; then
    BACKUP_DIR="/opt/ns/backups/postgres"
    mkdir -p "$BACKUP_DIR"
    chown -R postgres:postgres /opt/ns/backups

    cat << 'BACKUP_SCRIPT' > "$APP_DIR/backup_postgres.sh"
#!/bin/bash
# Automated PostgreSQL backup script for Neighbor Service
DB_NAME="REPLACE_DB_NAME"
BACKUP_DIR="/opt/ns/backups/postgres"
DATE=$(date +"%Y%m%d_%H%M%S")
RETENTION_DAYS=30
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${DATE}.sql.gz"

mkdir -p "$BACKUP_DIR"
sudo -u postgres pg_dump "$DB_NAME" | gzip > "$BACKUP_FILE"
find "$BACKUP_DIR" -name "*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete
echo "[$(date)] Backup completed: $BACKUP_FILE"
BACKUP_SCRIPT

    sed -i "s/REPLACE_DB_NAME/$DB_NAME/g" "$APP_DIR/backup_postgres.sh"
    chmod +x "$APP_DIR/backup_postgres.sh"
    
    (crontab -l 2>/dev/null | grep -v "backup_postgres.sh"; echo "0 2 * * * $APP_DIR/backup_postgres.sh >> $APP_DIR/logs/backup.log 2>&1") | crontab -
    echo -e "${GREEN}✓ Automated daily backup configured (02:00 AM UTC)${NC}"
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
    chown $USER:$GROUP "$APP_DIR/.env"
    chmod 600 "$APP_DIR/.env"
    echo -e "${CYAN}Please verify database credentials & secrets in $APP_DIR/.env${NC}"
else
    echo -e "${GREEN}✓ Existing .env preserved at $APP_DIR/.env${NC}"
    chmod 600 "$APP_DIR/.env"
fi

echo -e "${YELLOW}Phase 4: Direct Compilation (Go Binary Build)${NC}"

export PATH=$PATH:/usr/local/go/bin

echo -e "${YELLOW}Downloading Go dependencies...${NC}"
sudo -u $USER bash -c "cd '$APP_DIR' && env PATH=\"\$PATH:/usr/local/go/bin\" go mod tidy"

echo -e "${YELLOW}Compiling Neighbor Service API Server (cmd/server/main.go)...${NC}"
sudo -u $USER bash -c "cd '$APP_DIR' && env PATH=\"\$PATH:/usr/local/go/bin\" go build -ldflags='-s -w' -o bin/server ./cmd/server/"
chmod +x "$APP_DIR/bin/server"
echo -e "${GREEN}✓ Production API server compiled: $APP_DIR/bin/server${NC}"

echo -e "${YELLOW}Compiling Admin Generator CLI (cmd/create_admin/main.go)...${NC}"
sudo -u $USER bash -c "cd '$APP_DIR' && env PATH=\"\$PATH:/usr/local/go/bin\" go build -ldflags='-s -w' -o bin/create_admin ./cmd/create_admin/"
chmod +x "$APP_DIR/bin/create_admin"
echo -e "${GREEN}✓ Admin generator CLI compiled: $APP_DIR/bin/create_admin${NC}"

# Superuser / Administrator Account Creation Prompt
echo ""
echo -e "${YELLOW}Do you want to create or update an Admin Account now?${NC}"
echo "   1) Yes - Run Admin Account Creator"
echo "   2) No  - Skip admin creation"
read -p "   Enter choice [1-2, default 2]: " ADMIN_CHOICE
ADMIN_CHOICE=${ADMIN_CHOICE:-2}

if [ "$ADMIN_CHOICE" = "1" ]; then
    read -p "Enter Admin Email [default: admin@neighborservice.com]: " ADMIN_EMAIL
    ADMIN_EMAIL=${ADMIN_EMAIL:-admin@neighborservice.com}
    read -s -p "Enter Admin Password [default: Admin123!]: " ADMIN_PASS
    ADMIN_PASS=${ADMIN_PASS:-Admin123!}
    echo ""
    read -p "Enter First Name [default: Super]: " ADMIN_FN
    ADMIN_FN=${ADMIN_FN:-Super}
    read -p "Enter Last Name [default: Admin]: " ADMIN_LN
    ADMIN_LN=${ADMIN_LN:-Admin}
    
    sudo -u $USER env PATH="$PATH" "$APP_DIR/bin/create_admin" \
      -email="$ADMIN_EMAIL" \
      -password="$ADMIN_PASS" \
      -first-name="$ADMIN_FN" \
      -last-name="$ADMIN_LN" || echo -e "${YELLOW}Admin creation step completed.${NC}"
fi

echo -e "${YELLOW}Phase 5: Systemd Multi-Instance Cluster Setup${NC}"

mkdir -p "$APP_DIR/config"
chown -R $USER:$GROUP "$APP_DIR/config"

# 1. Template Unit File
cat << 'EOF' > /etc/systemd/system/ns-backend@.service
[Unit]
Description=Neighbor Service (NSA) Backend Instance #%i
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=APP_USER
Group=APP_GROUP
WorkingDirectory=APP_DIR_PATH
EnvironmentFile=APP_DIR_PATH/.env
EnvironmentFile=APP_DIR_PATH/config/instance-%i.env
ExecStart=APP_DIR_PATH/bin/server
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

sed -i "s|APP_USER|$USER|g" /etc/systemd/system/ns-backend@.service
sed -i "s|APP_GROUP|$GROUP|g" /etc/systemd/system/ns-backend@.service
sed -i "s|APP_DIR_PATH|$APP_DIR|g" /etc/systemd/system/ns-backend@.service

# 2. Master Cluster Target File
WANTS_LIST=""
for ((i=1; i<=NUM_INSTANCES; i++)); do
    WANTS_LIST="$WANTS_LIST ns-backend@$i.service"
done

cat << EOF > /etc/systemd/system/ns-backend.target
[Unit]
Description=Neighbor Service (NSA) Backend Multi-Instance Target
Wants=$WANTS_LIST

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload

# 3. Create per-instance env configs and start all configured instances
echo -e "${YELLOW}Starting $NUM_INSTANCES backend cluster instances...${NC}"
for ((i=1; i<=NUM_INSTANCES; i++)); do
    INSTANCE_PORT=$(($BASE_PORT + $i))
    cat << ENV_EOF > "$APP_DIR/config/instance-$i.env"
PORT=$INSTANCE_PORT
ENV_EOF
    chown $USER:$GROUP "$APP_DIR/config/instance-$i.env"
    chmod 600 "$APP_DIR/config/instance-$i.env"

    systemctl enable ns-backend@$i
    systemctl restart ns-backend@$i
    echo -e "${GREEN}  ✓ Instance #$i active on Port $INSTANCE_PORT${NC}"
done

# Stop any old higher instances if reducing count
for ((i=NUM_INSTANCES+1; i<=20; i++)); do
    if systemctl is-active --quiet ns-backend@$i; then
        systemctl stop ns-backend@$i || true
        systemctl disable ns-backend@$i || true
    fi
done

systemctl enable ns-backend.target
systemctl restart ns-backend.target
echo -e "${GREEN}✓ All $NUM_INSTANCES backend instances running and managed by ns-backend.target${NC}"

echo -e "${YELLOW}Phase 6: Nginx Load Balancer & SSL Gateway Configuration${NC}"

read -p "Do you want to configure Nginx Load Balancer with SSL for the API Gateway? (y/n) [default: y]: " SETUP_SSL
SETUP_SSL=${SETUP_SSL:-y}

# Generate Upstream Servers list
UPSTREAM_SERVERS=""
UPSTREAM_WS_SERVERS=""
for ((i=1; i<=NUM_INSTANCES; i++)); do
    PORT=$(($BASE_PORT + $i))
    UPSTREAM_SERVERS="${UPSTREAM_SERVERS}    server 127.0.0.1:${PORT} max_fails=3 fail_timeout=10s;\n"
    UPSTREAM_WS_SERVERS="${UPSTREAM_WS_SERVERS}    server 127.0.0.1:${PORT};\n"
done

if [ "$SETUP_SSL" = "y" ]; then
    read -p "Enter your API Domain (e.g. $DEFAULT_DOMAIN): " API_DOMAIN
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
# ------------------------------------------------------------------------------
# Nginx Upstream Load Balancer Pool: $NUM_INSTANCES Instances (Ports $(($BASE_PORT + 1)) - $(($BASE_PORT + $NUM_INSTANCES)))
# ------------------------------------------------------------------------------
upstream ns_backend_cluster {
    least_conn;
$(echo -e "$UPSTREAM_SERVERS")
    keepalive 64;
}

upstream ns_backend_ws_cluster {
    ip_hash;
$(echo -e "$UPSTREAM_WS_SERVERS")
}

# Rate limiting zones
limit_req_zone \$binary_remote_addr zone=ns_api_limit:10m rate=60r/s;
limit_req_zone \$binary_remote_addr zone=ns_auth_limit:10m rate=15r/s;

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

    # Rate-limited Auth Endpoints Load Balanced
    location /api/v1/accounts/login/ {
        limit_req zone=ns_auth_limit burst=20 nodelay;
        proxy_pass http://ns_backend_cluster;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
    }

    # Real-Time WebSocket Endpoint (Sticky via ip_hash)
    location /ws {
        proxy_pass http://ns_backend_ws_cluster/ws;
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

    # API Proxy to Load Balanced Upstream Cluster
    location / {
        limit_req zone=ns_api_limit burst=80 nodelay;
        proxy_pass http://ns_backend_cluster;
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
upstream ns_backend_cluster {
    least_conn;
$(echo -e "$UPSTREAM_SERVERS")
    keepalive 64;
}

server {
    listen 80;
    server_name _;
    server_tokens off;

    client_max_body_size 50M;

    location / {
        proxy_pass http://ns_backend_cluster;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "";
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
echo -e "${GREEN}✓ Nginx Load Balancer configured for $NUM_INSTANCES backend instances${NC}"

# Optional Cloudflare Cache Purge
if [ -n "$CF_ZONE_ID" ] && [ -n "$CF_API_TOKEN" ]; then
    echo -e "${YELLOW}Purging Cloudflare cache for API...${NC}"
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

echo -e "${YELLOW}Phase 7: Network Firewall (UFW) Hardening${NC}"
read -p "Do you want to configure and enable UFW firewall (allow 22, 80, 443; isolate backend ports & PostgreSQL)? (y/n) [default: y]: " SETUP_UFW
SETUP_UFW=${SETUP_UFW:-y}
if [ "$SETUP_UFW" = "y" ]; then
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow 22/tcp comment 'SSH'
    ufw allow 80/tcp comment 'HTTP'
    ufw allow 443/tcp comment 'HTTPS'
    ufw --force enable
    echo -e "${GREEN}✓ UFW Firewall enabled: ports 22, 80, 443 open. Backend instance ports $(($BASE_PORT + 1))-$(($BASE_PORT + $NUM_INSTANCES)) & PostgreSQL are isolated locally.${NC}"
fi

echo ""
echo -e "${BLUE}================================================================${NC}"
echo -e "${GREEN}✓ Production Backend Cluster Deployment Completed Successfully! ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}API Domain:       ${NC} https://$DEFAULT_DOMAIN"
echo -e "${CYAN}Active Instances: ${NC} $NUM_INSTANCES backend instances (Ports $(($BASE_PORT + 1)) to $(($BASE_PORT + $NUM_INSTANCES)))"
echo -e "${CYAN}Cluster Status:   ${NC} sudo systemctl status ns-backend.target"
echo -e "${CYAN}Single Instance:  ${NC} sudo systemctl status ns-backend@1"
echo -e "${CYAN}Cluster Logs:     ${NC} sudo journalctl -u 'ns-backend@*' -f"
echo ""
