#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER olx WITH PASSWORD 'olx';
    CREATE DATABASE mobile_numbers;
    GRANT ALL PRIVILEGES ON DATABASE mobile_numbers TO olx;
EOSQL