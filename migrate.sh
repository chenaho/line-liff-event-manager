#!/bin/bash

# ============================================================
# Firestore to PostgreSQL Migration Script
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${SCRIPT_DIR}/backend"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

show_help() {
    echo ""
    echo -e "${GREEN}Firestore to PostgreSQL Migration Script${NC}"
    echo ""
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  all         Migrate users, events, and interactions"
    echo "  users       Migrate only users"
    echo "  events      Migrate only events"
    echo "  interactions Migrate only interactions"
    echo "  verify      Verify migration counts"
    echo "  dry-run     Show what would be migrated"
    echo "  help        Show this help"
    echo ""
    echo "Environment Variables (set in .env or export):"
    echo "  FIREBASE_CREDENTIALS  Path to Firebase key file"
    echo "  POSTGRES_HOST         PostgreSQL host"
    echo "  POSTGRES_PORT         PostgreSQL port (default: 5433)"
    echo "  POSTGRES_USER         PostgreSQL user"
    echo "  POSTGRES_PASSWORD     PostgreSQL password"
    echo "  POSTGRES_DB           PostgreSQL database"
    echo ""
    echo "Example:"
    echo "  export FIREBASE_CREDENTIALS=./firebase-key.json"
    echo "  export POSTGRES_PASSWORD=your_password"
    echo "  $0 dry-run"
    echo "  $0 all"
    echo ""
}

check_env() {
    # Load .env if exists
    if [ -f "${SCRIPT_DIR}/.env" ]; then
        export $(grep -v '^#' "${SCRIPT_DIR}/.env" | xargs)
    fi

    if [ -z "$FIREBASE_CREDENTIALS" ]; then
        if [ -f "${SCRIPT_DIR}/firebase-key.json" ]; then
            # Use absolute path
            export FIREBASE_CREDENTIALS="$(cd "${SCRIPT_DIR}" && pwd)/firebase-key.json"
        else
            echo -e "${RED}Error: FIREBASE_CREDENTIALS not set${NC}"
            echo "Set it or place firebase-key.json in project root"
            exit 1
        fi
    else
        # Convert to absolute path if relative
        if [[ ! "$FIREBASE_CREDENTIALS" = /* ]]; then
            export FIREBASE_CREDENTIALS="$(cd "${SCRIPT_DIR}" && pwd)/${FIREBASE_CREDENTIALS}"
        fi
    fi

    if [ -z "$POSTGRES_PASSWORD" ]; then
        echo -e "${RED}Error: POSTGRES_PASSWORD not set${NC}"
        exit 1
    fi

    # Set defaults
    export POSTGRES_HOST=${POSTGRES_HOST:-localhost}
    export POSTGRES_PORT=${POSTGRES_PORT:-5433}
    export POSTGRES_USER=${POSTGRES_USER:-eventmanager}
    export POSTGRES_DB=${POSTGRES_DB:-eventmanager}
    
    echo -e "${BLUE}Firebase credentials: ${FIREBASE_CREDENTIALS}${NC}"
}


run_migration() {
    local command=$1
    
    echo -e "${BLUE}Building migration tool...${NC}"
    cd "${BACKEND_DIR}"
    go build -o migrate_tool ./cmd/migrate/migrate.go
    
    echo -e "${BLUE}Running migration: ${command}${NC}"
    ./migrate_tool "$command"
    
    rm -f migrate_tool
}

# Main
case "${1:-help}" in
    all)
        check_env
        run_migration "migrate-all"
        ;;
    users)
        check_env
        run_migration "migrate-users"
        ;;
    events)
        check_env
        run_migration "migrate-events"
        ;;
    interactions)
        check_env
        run_migration "migrate-interactions"
        ;;
    verify)
        check_env
        run_migration "verify"
        ;;
    dry-run)
        check_env
        run_migration "dry-run"
        ;;
    help|*)
        show_help
        ;;
esac
