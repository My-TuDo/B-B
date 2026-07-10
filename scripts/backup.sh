#!/bin/bash
# B-B Backup Script
# Usage: ./scripts/backup.sh
set -e

BACKUP_DIR="./backups"
DATE=$(date +"%Y%m%d_%H%M%S")

echo "=== B-B Backup ${DATE} ==="

# MySQL
echo "[MySQL] Dumping database..."
mkdir -p "${BACKUP_DIR}/mysql/daily"
docker exec bb-mysql mysqldump -u bb_user -pbb_password bb | gzip > "${BACKUP_DIR}/mysql/daily/bb_${DATE}.sql.gz"
# Keep last 7 days, remove older
find "${BACKUP_DIR}/mysql/daily" -name "bb_*.sql.gz" -mtime +7 -delete
echo "[MySQL] Done"

# Redis
echo "[Redis] Saving RDB..."
docker exec bb-redis redis-cli SAVE
mkdir -p "${BACKUP_DIR}/redis"
docker cp bb-redis:/data/dump.rdb "${BACKUP_DIR}/redis/dump_${DATE}.rdb"
# Keep last 7
find "${BACKUP_DIR}/redis" -name "dump_*.rdb" -mtime +7 -delete
echo "[Redis] Done"

# MinIO
echo "[MinIO] Mirroring buckets..."
mkdir -p "${BACKUP_DIR}/minio"
docker exec bb-minio-mc mc mirror myminio/bb-videos "/backups/minio/bb-videos_${DATE}" 2>/dev/null || \
  echo "[MinIO] Skipped (mc not available, manual backup required)"
# Keep last 7
find "${BACKUP_DIR}/minio" -name "bb-videos_*" -type d -mtime +7 -exec rm -rf {} + 2>/dev/null || true
echo "[MinIO] Done"

echo "=== Backup Complete ==="
