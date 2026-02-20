#!/bin/bash
set -e

DATA_DIR="${AGENTSKILLS_DATA_DIR:-/var/lib/agentskills}"
PG_DATA="$DATA_DIR/postgresql"
MINIO_DATA="$DATA_DIR/minio"
LOG_DIR="/var/log/agentskills"

mkdir -p "$PG_DATA" "$MINIO_DATA" "$LOG_DIR"

# ============================================
# 1. Initialize PostgreSQL if needed
# ============================================
if [ ! -f "$PG_DATA/PG_VERSION" ]; then
    echo "==> Initializing PostgreSQL database..."
    chown -R postgres:postgres "$PG_DATA"
    su postgres -c "initdb -D $PG_DATA --auth=trust --encoding=UTF8 --locale=C"

    # Configure pg_hba for local trust auth
    cat > "$PG_DATA/pg_hba.conf" <<PGHBA
local   all   all                 trust
host    all   all   127.0.0.1/32  trust
host    all   all   ::1/128       trust
PGHBA

    # Start temporarily to create DB and run init.sql
    su postgres -c "pg_ctl -D $PG_DATA -l $LOG_DIR/pg_init.log start -w"

    su postgres -c "createuser -s dev" || true
    su postgres -c "createdb -O dev agentskills" || true
    su postgres -c "psql -d agentskills -f /opt/agentskills/init.sql"

    su postgres -c "pg_ctl -D $PG_DATA stop -w"
    echo "==> PostgreSQL initialized."
else
    chown -R postgres:postgres "$PG_DATA"
    echo "==> PostgreSQL data directory exists, skipping init."
fi

# ============================================
# 2. Initialize MinIO data directory
# ============================================
chown -R minio-user:minio-user "$MINIO_DATA"

# ============================================
# 3. Start all services via supervisord
# ============================================
echo "==> Starting all services..."
exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
