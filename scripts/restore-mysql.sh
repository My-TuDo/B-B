#!/bin/bash
# B-B MySQL Restore Script
# Usage: ./scripts/restore-mysql.sh <backup_file>
set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <backup_file>"
  echo "Example: $0 ./backups/mysql/daily/bb_20260710_030000.sql.gz"
  exit 1
fi

BACKUP_FILE=$1

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Error: File not found: $BACKUP_FILE"
  exit 1
fi

echo "=== B-B MySQL Restore ==="
echo "Restoring from: $BACKUP_FILE"
echo "WARNING: This will overwrite the current database!"
read -p "Are you sure? (y/N): " confirm
if [ "$confirm" != "y" ]; then
  echo "Cancelled"
  exit 1
fi

echo "[MySQL] Restoring..."
if [[ "$BACKUP_FILE" == *.gz ]]; then
  gunzip -c "$BACKUP_FILE" | docker exec -i bb-mysql mysql -u bb_user -pbb_password bb
else
  docker exec -i bb-mysql mysql -u bb_user -pbb_password bb < "$BACKUP_FILE"
fi
echo "[MySQL] Restore complete"
