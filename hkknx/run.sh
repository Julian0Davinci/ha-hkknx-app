#!/bin/sh
set -eu

DB_DIR=/data/hkknx
TRANSFER_DIR=/share/hkknx-migration
# Nie eine neue HomeKit-Identitaet anlegen, wenn noch keine Daten importiert wurden.
if [ ! -d "$DB_DIR" ] || [ -z "$(ls -A "$DB_DIR" 2>/dev/null || true)" ]; then
    if [ ! -f "$TRANSFER_DIR/.ready" ]; then
        echo "HKKNX: STOP - noch keine Daten unter $DB_DIR und kein Migrationsexport unter $TRANSFER_DIR/.ready"
        echo "HKKNX: Erst Alt-App sichern/exportieren. Die neue App wird NICHT mit leerer HomeKit-DB gestartet."
        exit 1
    fi
    mkdir -p "$DB_DIR"
    chmod 700 "$DB_DIR"
    cp -a "$TRANSFER_DIR/." "$DB_DIR/"
    rm -f "$DB_DIR/.ready"
    touch "$DB_DIR/.imported_from_local_hkknx"
    echo "HKKNX: bestehende Konfiguration in privates persistentes /data importiert."
fi

# Naechste Starts und Updates verwenden das private Volume; kein Re-Import.
echo "HKKNX: start database=$DB_DIR, port=8080"
exec hkknx --db "$DB_DIR" --port 8080
