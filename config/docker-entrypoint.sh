#!/bin/sh
set -e

mkdir -p /data/uploads

if [ "$(stat -c '%u' /data)" != "1001" ] || [ ! -w /data ]; then
    chown -R app:app /data
fi

exec su-exec app "$@"
