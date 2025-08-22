#!/bin/sh

set -e

echo "Running migrations..."
goose -dir /migrations postgres "$POSTGRES_DSN" up

echo "Migrations completed"