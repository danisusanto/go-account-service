#!/bin/sh
set -e

echo "Creating replication role if not exists..."

psql -U "$POSTGRES_USER" -d postgres <<-EOSQL
DO
\$\$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_roles WHERE rolname = 'replicator'
   ) THEN
      CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'secret';
   END IF;
END
\$\$;
EOSQL

echo "Replication role ready."