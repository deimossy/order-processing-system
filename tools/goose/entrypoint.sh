#!/bin/sh

set -e

echo "Running migrations..."
goose -dir /migrations postgres "$DB_URL" up

echo "Migrations completed"