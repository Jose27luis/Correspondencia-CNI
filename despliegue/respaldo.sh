#!/bin/bash
set -euo pipefail

DESTINO="/root/backups-correspondencia"
RETENCION_DIAS=14
FECHA=$(date +%Y%m%d-%H%M%S)
ARCHIVO="$DESTINO/correspondencia-$FECHA.sql.gz"

mkdir -p "$DESTINO"

URL=$(grep -m1 '^DATABASE_URL=' /var/www/Correspondencia-CNI/backend/.env | cut -d= -f2-)

if [ -z "$URL" ]; then
    echo "no se encontró DATABASE_URL" >&2
    exit 1
fi

pg_dump "$URL" | gzip > "$ARCHIVO"

if [ ! -s "$ARCHIVO" ]; then
    echo "el respaldo quedó vacío" >&2
    rm -f "$ARCHIVO"
    exit 1
fi

if ! gzip -t "$ARCHIVO"; then
    echo "el respaldo está corrupto" >&2
    rm -f "$ARCHIVO"
    exit 1
fi

find "$DESTINO" -name 'correspondencia-*.sql.gz' -mtime "+$RETENCION_DIAS" -delete

echo "respaldo creado: $ARCHIVO ($(du -h "$ARCHIVO" | cut -f1))"
