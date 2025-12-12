#!/bin/bash

# ============================================================
# Event Manager Backend Deployment Script
# Supports both Firestore and PostgreSQL databases
# ============================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect docker compose command (docker-compose vs docker compose)
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    echo -e "${RED}Error: Neither 'docker-compose' nor 'docker compose' found!${NC}"
    exit 1
fi

echo -e "${BLUE}Using: ${DOCKER_COMPOSE}${NC}"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.prod.yml"

# Functions
show_help() {
    echo ""
    echo -e "${GREEN}Event Manager Backend Deployment Script${NC}"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  deploy    Deploy the backend (default)"
    echo "  stop      Stop all services"
    echo "  restart   Restart all services"
    echo "  logs      Show logs"
    echo "  status    Show container status"
    echo "  help      Show this help message"
    echo ""
    echo "Options:"
    echo "  --db=firestore    Use Firestore database (default)"
    echo "  --db=postgres     Use PostgreSQL database"
    echo "  --build           Force rebuild containers"
    echo "  --no-cache        Build without cache"
    echo ""
    echo "Examples:"
    echo "  $0 deploy --db=firestore"
    echo "  $0 deploy --db=postgres --build"
    echo "  $0 logs"
    echo "  $0 stop"
    echo ""
}

check_env_file() {
    if [ ! -f "${SCRIPT_DIR}/.env" ]; then
        echo -e "${YELLOW}Warning: .env file not found. Creating from template...${NC}"
        cat > "${SCRIPT_DIR}/.env" << 'EOF'
# Database Selection
DB_TYPE=firestore

# JWT Secret (change in production!)
JWT_SECRET=your-jwt-secret-change-this

# LINE Configuration
LINE_CHANNEL_ID=your-line-channel-id
ADMIN_LIST=user1,user2

# PostgreSQL Configuration (only needed if DB_TYPE=postgres)
POSTGRES_USER=eventmanager
POSTGRES_PASSWORD=changeme
POSTGRES_DB=eventmanager
EOF
        echo -e "${YELLOW}Please edit .env file with your configuration!${NC}"
    fi
}

deploy_firestore() {
    echo -e "${GREEN}Deploying with Firestore...${NC}"
    
    # Check for firebase key
    if [ ! -f "${SCRIPT_DIR}/firebase-key.json" ]; then
        echo -e "${RED}Error: firebase-key.json not found!${NC}"
        echo "Please place your Firebase service account key in: ${SCRIPT_DIR}/firebase-key.json"
        exit 1
    fi
    
    # Set DB_TYPE in .env
    sed -i.bak 's/^DB_TYPE=.*/DB_TYPE=firestore/' "${SCRIPT_DIR}/.env" 2>/dev/null || \
    sed -i '' 's/^DB_TYPE=.*/DB_TYPE=firestore/' "${SCRIPT_DIR}/.env"
    
    # Deploy without postgres profile (Caddy managed externally via web_gateway)
    if [ "$BUILD_FLAG" = "true" ]; then
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" up -d --build $NO_CACHE_FLAG app-backend
    else
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" up -d app-backend
    fi
    
    echo -e "${GREEN}✓ Deployed with Firestore${NC}"
}


deploy_postgres() {
    echo -e "${GREEN}Deploying with PostgreSQL...${NC}"
    
    # Set DB_TYPE in .env
    sed -i.bak 's/^DB_TYPE=.*/DB_TYPE=postgres/' "${SCRIPT_DIR}/.env" 2>/dev/null || \
    sed -i '' 's/^DB_TYPE=.*/DB_TYPE=postgres/' "${SCRIPT_DIR}/.env"
    
    # Deploy with postgres profile
    if [ "$BUILD_FLAG" = "true" ]; then
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --profile postgres up -d --build $NO_CACHE_FLAG
    else
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --profile postgres up -d
    fi
    
    echo -e "${GREEN}✓ Deployed with PostgreSQL${NC}"
    echo -e "${BLUE}PostgreSQL is available on port 5433${NC}"
}

stop_services() {
    echo -e "${YELLOW}Stopping all services...${NC}"
    $DOCKER_COMPOSE -f "$COMPOSE_FILE" --profile postgres down
    echo -e "${GREEN}✓ All services stopped${NC}"
}

show_logs() {
    $DOCKER_COMPOSE -f "$COMPOSE_FILE" logs -f --tail=100
}

show_status() {
    echo -e "${BLUE}Container Status:${NC}"
    $DOCKER_COMPOSE -f "$COMPOSE_FILE" --profile postgres ps
}

# Parse arguments
COMMAND="deploy"
DB_TYPE="firestore"
BUILD_FLAG="false"
NO_CACHE_FLAG=""

for arg in "$@"; do
    case $arg in
        deploy|stop|restart|logs|status|help)
            COMMAND="$arg"
            ;;
        --db=*)
            DB_TYPE="${arg#*=}"
            ;;
        --build)
            BUILD_FLAG="true"
            ;;
        --no-cache)
            NO_CACHE_FLAG="--no-cache"
            BUILD_FLAG="true"
            ;;
        *)
            echo -e "${RED}Unknown option: $arg${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Execute command
case $COMMAND in
    deploy)
        check_env_file
        case $DB_TYPE in
            firestore)
                deploy_firestore
                ;;
            postgres)
                deploy_postgres
                ;;
            *)
                echo -e "${RED}Unknown database type: $DB_TYPE${NC}"
                echo "Supported types: firestore, postgres"
                exit 1
                ;;
        esac
        echo ""
        show_status
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        check_env_file
        case $DB_TYPE in
            firestore)
                deploy_firestore
                ;;
            postgres)
                deploy_postgres
                ;;
        esac
        ;;
    logs)
        show_logs
        ;;
    status)
        show_status
        ;;
    help)
        show_help
        ;;
esac
