#!/bin/sh
# Creates additional databases needed by non-default services.
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    SELECT 'CREATE DATABASE restaurantdb'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'restaurantdb')\gexec
EOSQL
