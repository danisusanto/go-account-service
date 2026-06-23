#!/bin/sh
set -e

PGDATA=/var/lib/postgresql/data

echo "Waiting for master..."

until pg_isready -h go-account-service-db -p 5432 -U postgres
do
  sleep 2
done

echo "Cleaning data directory..."
rm -rf $PGDATA/*

echo "Starting base backup..."

PGPASSWORD=secret pg_basebackup \
  -h go-account-service-db \
  -D $PGDATA \
  -U replicator \
  -Fp -Xs -P -R

echo "Fixing permissions..."
chown -R postgres:postgres $PGDATA

echo "Starting postgres replica..."
exec gosu postgres postgres
