#!/bin/sh
set -eu

# Supervisor mounts this app's own persistent /data folder.
DB_DIR=/data/hkknx
mkdir -p "$DB_DIR"
chmod 700 "$DB_DIR"

# A separate proxy serves ONLY Home Assistant Ingress on port 8099.
# HKKNX itself stays on its unchanged direct-access port 8080.
/usr/local/bin/hkknx-ingress-proxy &
echo "HKKNX: ingress proxy launched (port 8099)"
echo "HKKNX: database=$DB_DIR, direct web UI port=8080"
exec hkknx --db "$DB_DIR" --port 8080
