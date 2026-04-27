#!/bin/sh
set -e

echo "Convoy API - Starting..."

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; do
  echo "  PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready!"

# Run migrations
echo "Running database migrations..."
DATABASE_URL="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE"

if [ -d "./migrations" ]; then
  # Try to run migrations
  migrate -path ./migrations -database "$DATABASE_URL" up 2>&1 | tee /tmp/migrate.log
  MIGRATE_EXIT_CODE=$?
  
  # Check if migration failed due to dirty state
  if grep -q "Dirty database version" /tmp/migrate.log; then
    echo "WARNING: Detected dirty migration state!"
    echo "   This usually means a previous migration failed partway through."
    echo "   Resetting migrations to clean state..."
    
    # Drop all migrations and start fresh
    echo "   Dropping all migrations..."
    migrate -path ./migrations -database "$DATABASE_URL" drop -f
    
    # Run migrations from scratch
    echo "   Running migrations from scratch..."
    migrate -path ./migrations -database "$DATABASE_URL" up
    
    if [ $? -eq 0 ]; then
      echo "SUCCESS: Migrations completed successfully after reset!"
    else
      echo "ERROR: Migrations failed after reset!"
      echo "   Check the migration files for errors."
      exit 1
    fi
  elif [ $MIGRATE_EXIT_CODE -eq 0 ]; then
    echo "SUCCESS: Migrations completed successfully!"
  else
    echo "ERROR: Migrations failed!"
    echo "   Error details:"
    cat /tmp/migrate.log
    exit 1
  fi
else
  echo "WARNING: No migrations directory found, skipping migrations"
fi

echo "Starting Convoy API server..."
exec "$@"
