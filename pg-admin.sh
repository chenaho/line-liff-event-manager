#!/bin/bash

# ============================================================
# Event Manager PostgreSQL Management Script
# Operations for PostgreSQL database on Linux server
# ============================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="event-postgres"
DB_NAME="${POSTGRES_DB:-eventmanager}"
DB_USER="${POSTGRES_USER:-eventmanager}"
DB_PORT="${POSTGRES_PORT:-5433}"

# Detect docker compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE=""
fi

# Functions
show_help() {
    echo ""
    echo -e "${GREEN}Event Manager PostgreSQL Management Script${NC}"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  status        Show PostgreSQL container status"
    echo "  connect       Connect to PostgreSQL shell (psql)"
    echo "  logs          Show PostgreSQL logs"
    echo "  backup        Backup database to file"
    echo "  restore       Restore database from file"
    echo "  init          Initialize/reset database schema"
    echo "  query         Execute SQL query"
    echo "  tables        List all tables"
    echo "  count         Show record counts"
    echo "  export        Export table to CSV"
    echo "  import        Import CSV to table"
    echo "  health        Check database health"
    echo "  vacuum        Run VACUUM ANALYZE"
    echo "  reset         Reset database (WARNING: deletes all data!)"
    echo "  help          Show this help message"
    echo ""
    echo "Options:"
    echo "  -f, --file    File path for backup/restore/import/export"
    echo "  -t, --table   Table name for export/import"
    echo "  -q, --query   SQL query to execute"
    echo ""
    echo "Examples:"
    echo "  $0 status"
    echo "  $0 connect"
    echo "  $0 backup -f ./backup.sql"
    echo "  $0 restore -f ./backup.sql"
    echo "  $0 query -q 'SELECT count(*) FROM events'"
    echo "  $0 export -t events -f ./events.csv"
    echo "  $0 tables"
    echo ""
}

check_container() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "${RED}Error: PostgreSQL container '${CONTAINER_NAME}' is not running!${NC}"
        echo "Start it with: ./deploy.sh deploy --db=postgres"
        exit 1
    fi
}

show_status() {
    echo -e "${BLUE}PostgreSQL Container Status:${NC}"
    docker ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "${GREEN}✓ PostgreSQL is running${NC}"
        echo ""
        echo -e "${CYAN}Connection Info:${NC}"
        echo "  Host: localhost"
        echo "  Port: ${DB_PORT}"
        echo "  User: ${DB_USER}"
        echo "  Database: ${DB_NAME}"
    else
        echo -e "${RED}✗ PostgreSQL is not running${NC}"
    fi
}

connect_shell() {
    check_container
    echo -e "${BLUE}Connecting to PostgreSQL shell...${NC}"
    docker exec -it ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME}
}

show_logs() {
    docker logs -f ${CONTAINER_NAME} --tail=100
}

backup_database() {
    check_container
    local backup_file="${FILE_PATH:-./backup_$(date +%Y%m%d_%H%M%S).sql}"
    
    echo -e "${BLUE}Backing up database to: ${backup_file}${NC}"
    docker exec ${CONTAINER_NAME} pg_dump -U ${DB_USER} -d ${DB_NAME} > "${backup_file}"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Backup completed: ${backup_file}${NC}"
        echo "  Size: $(ls -lh "${backup_file}" | awk '{print $5}')"
    else
        echo -e "${RED}✗ Backup failed${NC}"
        exit 1
    fi
}

restore_database() {
    check_container
    
    if [ -z "$FILE_PATH" ]; then
        echo -e "${RED}Error: Please specify backup file with -f option${NC}"
        exit 1
    fi
    
    if [ ! -f "$FILE_PATH" ]; then
        echo -e "${RED}Error: File not found: ${FILE_PATH}${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}WARNING: This will overwrite existing data!${NC}"
    read -p "Are you sure? (y/N): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        echo "Cancelled."
        exit 0
    fi
    
    echo -e "${BLUE}Restoring database from: ${FILE_PATH}${NC}"
    cat "${FILE_PATH}" | docker exec -i ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME}
    
    echo -e "${GREEN}✓ Restore completed${NC}"
}

init_schema() {
    check_container
    local init_file="./backend/init.sql"
    
    if [ ! -f "$init_file" ]; then
        echo -e "${RED}Error: init.sql not found at ${init_file}${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}Initializing database schema...${NC}"
    cat "${init_file}" | docker exec -i ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME}
    
    echo -e "${GREEN}✓ Schema initialized${NC}"
}

execute_query() {
    check_container
    
    if [ -z "$QUERY" ]; then
        echo -e "${RED}Error: Please specify query with -q option${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}Executing query...${NC}"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "${QUERY}"
}

list_tables() {
    check_container
    echo -e "${BLUE}Tables in database:${NC}"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\dt"
}

show_counts() {
    check_container
    echo -e "${BLUE}Record counts:${NC}"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "
        SELECT 'events' as table_name, COUNT(*) as count FROM events
        UNION ALL
        SELECT 'interactions', COUNT(*) FROM interactions
        UNION ALL
        SELECT 'users', COUNT(*) FROM users
        ORDER BY table_name;
    "
}

export_table() {
    check_container
    
    if [ -z "$TABLE_NAME" ]; then
        echo -e "${RED}Error: Please specify table with -t option${NC}"
        exit 1
    fi
    
    local export_file="${FILE_PATH:-./export_${TABLE_NAME}_$(date +%Y%m%d_%H%M%S).csv}"
    
    echo -e "${BLUE}Exporting table '${TABLE_NAME}' to: ${export_file}${NC}"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\COPY ${TABLE_NAME} TO STDOUT WITH CSV HEADER" > "${export_file}"
    
    echo -e "${GREEN}✓ Export completed: ${export_file}${NC}"
}

import_table() {
    check_container
    
    if [ -z "$TABLE_NAME" ]; then
        echo -e "${RED}Error: Please specify table with -t option${NC}"
        exit 1
    fi
    
    if [ -z "$FILE_PATH" ] || [ ! -f "$FILE_PATH" ]; then
        echo -e "${RED}Error: Please specify valid CSV file with -f option${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}Importing to table '${TABLE_NAME}' from: ${FILE_PATH}${NC}"
    cat "${FILE_PATH}" | docker exec -i ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\COPY ${TABLE_NAME} FROM STDIN WITH CSV HEADER"
    
    echo -e "${GREEN}✓ Import completed${NC}"
}

health_check() {
    check_container
    echo -e "${BLUE}Database Health Check:${NC}"
    echo ""
    
    echo "Connection test..."
    if docker exec ${CONTAINER_NAME} pg_isready -U ${DB_USER} -d ${DB_NAME}; then
        echo -e "${GREEN}✓ Connection OK${NC}"
    else
        echo -e "${RED}✗ Connection failed${NC}"
    fi
    echo ""
    
    echo "Database size..."
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "
        SELECT pg_size_pretty(pg_database_size('${DB_NAME}')) as database_size;
    "
    
    echo "Table sizes..."
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "
        SELECT 
            tablename as table,
            pg_size_pretty(pg_total_relation_size(schemaname || '.' || tablename)) as size
        FROM pg_tables 
        WHERE schemaname = 'public'
        ORDER BY pg_total_relation_size(schemaname || '.' || tablename) DESC;
    "
}

vacuum_analyze() {
    check_container
    echo -e "${BLUE}Running VACUUM ANALYZE...${NC}"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "VACUUM ANALYZE;"
    echo -e "${GREEN}✓ VACUUM ANALYZE completed${NC}"
}

reset_database() {
    check_container
    
    echo -e "${RED}WARNING: This will DELETE ALL DATA in the database!${NC}"
    read -p "Are you sure? Type 'YES' to confirm: " confirm
    if [ "$confirm" != "YES" ]; then
        echo "Cancelled."
        exit 0
    fi
    
    echo -e "${YELLOW}Resetting database...${NC}"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "
        DROP TABLE IF EXISTS interactions CASCADE;
        DROP TABLE IF EXISTS events CASCADE;
        DROP TABLE IF EXISTS users CASCADE;
    "
    
    init_schema
    
    echo -e "${GREEN}✓ Database reset completed${NC}"
}

# Parse arguments
COMMAND=""
FILE_PATH=""
TABLE_NAME=""
QUERY=""

while [[ $# -gt 0 ]]; do
    case $1 in
        status|connect|logs|backup|restore|init|query|tables|count|export|import|health|vacuum|reset|help)
            COMMAND="$1"
            shift
            ;;
        -f|--file)
            FILE_PATH="$2"
            shift 2
            ;;
        -t|--table)
            TABLE_NAME="$2"
            shift 2
            ;;
        -q|--query)
            QUERY="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Default command
if [ -z "$COMMAND" ]; then
    COMMAND="help"
fi

# Execute command
case $COMMAND in
    status)     show_status ;;
    connect)    connect_shell ;;
    logs)       show_logs ;;
    backup)     backup_database ;;
    restore)    restore_database ;;
    init)       init_schema ;;
    query)      execute_query ;;
    tables)     list_tables ;;
    count)      show_counts ;;
    export)     export_table ;;
    import)     import_table ;;
    health)     health_check ;;
    vacuum)     vacuum_analyze ;;
    reset)      reset_database ;;
    help)       show_help ;;
esac
