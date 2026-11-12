#!/bin/sh
set -eu

# The HA Supervisor supplies persistent /data, separate from the Docker image.
DB_DIR=/data/hkknx
mkdir -p "$DB_DIR"
chmod 700 "$DB_DIR"
echo "HKKNX: starting with persistent database at $DB_DIR"
exec hkknx --db "$DB_DIR" --port 8080
