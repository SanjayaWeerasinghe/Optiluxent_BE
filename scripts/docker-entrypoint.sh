#!/bin/sh
set -e

echo "==> Running database migrations..."
./erp-migrate -command up

echo "==> Starting ERP server..."
exec ./erp-server
