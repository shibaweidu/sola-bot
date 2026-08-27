#!/bin/sh
set -eu

: "${DATABASE_URL:?DATABASE_URL is required}"

psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c \
  "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"

for file in /migrations/*.up.sql; do
  [ -f "$file" ] || continue

  filename="$(basename "$file")"
  version="${filename%%_*}"
  case "$version" in
    ''|*[!0-9]*)
      echo "Invalid migration filename: $filename" >&2
      exit 1
      ;;
  esac

  if [ "$(psql "$DATABASE_URL" -tAc "SELECT 1 FROM schema_migrations WHERE version = '$version'")" = "1" ]; then
    echo "Skipping migration $version"
    continue
  fi

  echo "Applying migration $version from $filename"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 --single-transaction \
    -f "$file" \
    -c "INSERT INTO schema_migrations (version) VALUES ('$version');"
done

echo "Database migrations are up to date."
