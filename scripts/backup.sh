#!/bin/bash

# ==========================================
# BOME Webapp Backup Script
# ==========================================
# 
# This script creates backups of the database and Docker volumes
# Run this script regularly via cron for automated backups
#
# Usage: ./scripts/backup.sh [options]
# Options:
#   -d, --database    Backup database only
#   -v, --volumes     Backup volumes only
#   -c, --cleanup     Cleanup old backups
#   -h, --help        Show this help message
#
# Version: 1.0.0
# Last Updated: $(date)

set -euo pipefail

# Configuration
BACKUP_DIR="./backups"
DB_NAME="bome_db"
DB_USER="bome_user"
RETENTION_DAYS=30
LOG_FILE="./logs/backup.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create necessary directories
mkdir -p "$BACKUP_DIR" "$(dirname "$LOG_FILE")"

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    echo -e "${RED}ERROR: $1${NC}" >&2
    log "ERROR: $1"
    exit 1
}

# Success message
success() {
    echo -e "${GREEN}SUCCESS: $1${NC}"
    log "SUCCESS: $1"
}

# Warning message
warning() {
    echo -e "${YELLOW}WARNING: $1${NC}"
    log "WARNING: $1"
}

# Info message
info() {
    echo -e "${BLUE}INFO: $1${NC}"
    log "INFO: $1"
}

# Check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        error_exit "docker-compose is not installed or not in PATH"
    fi
    
    if ! docker-compose ps &> /dev/null; then
        error_exit "Docker Compose services are not running"
    fi
}

# Check if PostgreSQL container is running
check_postgres() {
    if ! docker-compose exec postgres pg_isready -U "$DB_USER" &> /dev/null; then
        error_exit "PostgreSQL is not ready or not running"
    fi
}

# Database backup function
backup_database() {
    info "Starting database backup..."
    
    check_postgres
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/database_backup_${timestamp}.sql"
    local backup_file_compressed="$backup_file.gz"
    
    # Create database backup
    if docker-compose exec postgres pg_dump -U "$DB_USER" "$DB_NAME" > "$backup_file"; then
        # Compress the backup
        if gzip "$backup_file"; then
            local file_size=$(du -h "$backup_file_compressed" | cut -f1)
            success "Database backup created: $backup_file_compressed ($file_size)"
        else
            warning "Database backup created but compression failed: $backup_file"
        fi
    else
        error_exit "Failed to create database backup"
    fi
    
    # Create backup metadata
    cat > "$BACKUP_DIR/database_backup_${timestamp}.meta" << EOF
{
    "type": "database",
    "timestamp": "$timestamp",
    "database": "$DB_NAME",
    "user": "$DB_USER",
    "file": "$(basename "$backup_file_compressed")",
    "size": "$(du -b "$backup_file_compressed" | cut -f1)",
    "created_at": "$(date -Iseconds)"
}
EOF
}

# Volume backup function
backup_volumes() {
    info "Starting volume backup..."
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    # Backup PostgreSQL data
    local postgres_backup="$BACKUP_DIR/postgres_volume_${timestamp}.tar.gz"
    if docker run --rm -v bome_postgres_data:/data -v "$(pwd)/$BACKUP_DIR":/backup alpine tar czf "/backup/$(basename "$postgres_backup")" -C /data .; then
        local file_size=$(du -h "$postgres_backup" | cut -f1)
        success "PostgreSQL volume backup created: $postgres_backup ($file_size)"
    else
        error_exit "Failed to create PostgreSQL volume backup"
    fi
    
    # Backup Redis data
    local redis_backup="$BACKUP_DIR/redis_volume_${timestamp}.tar.gz"
    if docker run --rm -v bome_redis_data:/data -v "$(pwd)/$BACKUP_DIR":/backup alpine tar czf "/backup/$(basename "$redis_backup")" -C /data .; then
        local file_size=$(du -h "$redis_backup" | cut -f1)
        success "Redis volume backup created: $redis_backup ($file_size)"
    else
        warning "Failed to create Redis volume backup"
    fi
    
    # Create volume backup metadata
    cat > "$BACKUP_DIR/volume_backup_${timestamp}.meta" << EOF
{
    "type": "volume",
    "timestamp": "$timestamp",
    "volumes": [
        {
            "name": "postgres_data",
            "file": "$(basename "$postgres_backup")",
            "size": "$(du -b "$postgres_backup" | cut -f1)"
        },
        {
            "name": "redis_data",
            "file": "$(basename "$redis_backup")",
            "size": "$(du -b "$redis_backup" | cut -f1)"
        }
    ],
    "created_at": "$(date -Iseconds)"
}
EOF
}

# Cleanup old backups
cleanup_backups() {
    info "Cleaning up old backups (older than $RETENTION_DAYS days)..."
    
    local deleted_count=0
    
    # Find and delete old backup files
    while IFS= read -r -d '' file; do
        rm -f "$file"
        ((deleted_count++))
        info "Deleted old backup: $(basename "$file")"
    done < <(find "$BACKUP_DIR" -name "*.sql.gz" -o -name "*.tar.gz" -o -name "*.meta" -type f -mtime +$RETENTION_DAYS -print0)
    
    if [ $deleted_count -eq 0 ]; then
        info "No old backups to clean up"
    else
        success "Cleaned up $deleted_count old backup files"
    fi
}

# List existing backups
list_backups() {
    info "Listing existing backups..."
    
    if [ ! -d "$BACKUP_DIR" ] || [ -z "$(ls -A "$BACKUP_DIR" 2>/dev/null)" ]; then
        warning "No backups found in $BACKUP_DIR"
        return
    fi
    
    echo -e "\n${BLUE}Database Backups:${NC}"
    find "$BACKUP_DIR" -name "database_backup_*.sql.gz" -type f -exec ls -lh {} \; | sort -k6,7
    
    echo -e "\n${BLUE}Volume Backups:${NC}"
    find "$BACKUP_DIR" -name "*_volume_*.tar.gz" -type f -exec ls -lh {} \; | sort -k6,7
    
    echo -e "\n${BLUE}Disk Usage:${NC}"
    du -sh "$BACKUP_DIR"
}

# Restore database from backup
restore_database() {
    local backup_file="$1"
    
    if [ ! -f "$backup_file" ]; then
        error_exit "Backup file not found: $backup_file"
    fi
    
    warning "This will OVERWRITE the current database. Are you sure? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        info "Database restore cancelled"
        return
    fi
    
    info "Restoring database from: $backup_file"
    
    check_postgres
    
    # Decompress if needed
    local sql_file="$backup_file"
    if [[ "$backup_file" == *.gz ]]; then
        sql_file="${backup_file%.gz}"
        if ! gunzip -c "$backup_file" > "$sql_file"; then
            error_exit "Failed to decompress backup file"
        fi
    fi
    
    # Restore database
    if docker-compose exec postgres psql -U "$DB_USER" -d "$DB_NAME" < "$sql_file"; then
        success "Database restored successfully"
    else
        error_exit "Failed to restore database"
    fi
    
    # Clean up temporary file
    if [[ "$backup_file" == *.gz ]]; then
        rm -f "$sql_file"
    fi
}

# Verify backup integrity
verify_backup() {
    local backup_file="$1"
    
    if [ ! -f "$backup_file" ]; then
        error_exit "Backup file not found: $backup_file"
    fi
    
    info "Verifying backup integrity: $backup_file"
    
    if [[ "$backup_file" == *.gz ]]; then
        if gunzip -t "$backup_file"; then
            success "Backup file integrity verified"
        else
            error_exit "Backup file is corrupted"
        fi
    elif [[ "$backup_file" == *.sql ]]; then
        if head -n 5 "$backup_file" | grep -q "PostgreSQL database dump"; then
            success "Backup file appears to be valid"
        else
            error_exit "Backup file does not appear to be a valid PostgreSQL dump"
        fi
    else
        warning "Cannot verify backup file type: $backup_file"
    fi
}

# Show help
show_help() {
    cat << EOF
BOME Webapp Backup Script

Usage: $0 [OPTIONS]

OPTIONS:
    -d, --database              Backup database only
    -v, --volumes               Backup volumes only
    -c, --cleanup               Cleanup old backups
    -l, --list                  List existing backups
    -r, --restore <file>        Restore database from backup file
    -V, --verify <file>         Verify backup file integrity
    -h, --help                  Show this help message

EXAMPLES:
    $0                          # Full backup (database + volumes)
    $0 -d                       # Database backup only
    $0 -v                       # Volume backup only
    $0 -c                       # Cleanup old backups
    $0 -l                       # List existing backups
    $0 -r backup.sql.gz         # Restore from backup
    $0 -V backup.sql.gz         # Verify backup integrity

CONFIGURATION:
    Backup directory: $BACKUP_DIR
    Retention days: $RETENTION_DAYS
    Log file: $LOG_FILE

CRON EXAMPLE:
    # Daily backup at 2 AM
    0 2 * * * /path/to/backup.sh

    # Weekly cleanup on Sunday at 3 AM
    0 3 * * 0 /path/to/backup.sh -c

EOF
}

# Main function
main() {
    local database_only=false
    local volumes_only=false
    local cleanup_only=false
    local list_only=false
    local restore_file=""
    local verify_file=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--database)
                database_only=true
                shift
                ;;
            -v|--volumes)
                volumes_only=true
                shift
                ;;
            -c|--cleanup)
                cleanup_only=true
                shift
                ;;
            -l|--list)
                list_only=true
                shift
                ;;
            -r|--restore)
                restore_file="$2"
                shift 2
                ;;
            -V|--verify)
                verify_file="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                error_exit "Unknown option: $1"
                ;;
        esac
    done
    
    # Check Docker Compose availability
    check_docker_compose
    
    # Execute based on options
    if [ -n "$restore_file" ]; then
        restore_database "$restore_file"
    elif [ -n "$verify_file" ]; then
        verify_backup "$verify_file"
    elif [ "$list_only" = true ]; then
        list_backups
    elif [ "$cleanup_only" = true ]; then
        cleanup_backups
    elif [ "$database_only" = true ]; then
        backup_database
    elif [ "$volumes_only" = true ]; then
        backup_volumes
    else
        # Full backup (default)
        backup_database
        backup_volumes
        cleanup_backups
    fi
    
    success "Backup operations completed successfully"
}

# Run main function with all arguments
main "$@" 