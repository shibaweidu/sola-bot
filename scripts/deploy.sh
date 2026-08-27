#!/bin/sh
set -eu

unset CDPATH
script_dir="$(cd -- "$(dirname -- "$0")" && pwd)"
deploy_dir="$(dirname "$script_dir")"
compose_file="${SOLA_COMPOSE_FILE:-$deploy_dir/docker-compose.prod.yml}"
env_file="${SOLA_ENV_FILE:-$deploy_dir/.env}"

if [ ! -f "$compose_file" ]; then
  echo "Missing production compose file: $compose_file" >&2
  exit 1
fi
if [ ! -f "$env_file" ]; then
  echo "Missing server environment file: $env_file" >&2
  exit 1
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is not installed." >&2
  exit 1
fi

if [ "${1:-}" != "" ]; then
  SOLA_IMAGE_TAG="$1"
fi
export SOLA_IMAGE_REGISTRY="${SOLA_IMAGE_REGISTRY:-ghcr.io/shibaweidu/sola-bot}"
export SOLA_IMAGE_TAG="${SOLA_IMAGE_TAG:-latest}"

compose() {
  docker compose --env-file "$env_file" -f "$compose_file" "$@"
}

echo "Deploying $SOLA_IMAGE_REGISTRY with tag $SOLA_IMAGE_TAG"
compose config >/dev/null
compose pull
compose up -d --remove-orphans

services="postgres redis api bot worker nginx"
timeout_seconds="${SOLA_DEPLOY_TIMEOUT:-180}"
elapsed=0

while [ "$elapsed" -lt "$timeout_seconds" ]; do
  ready=true
  for service in $services; do
    container_id="$(compose ps -q "$service")"
    if [ -z "$container_id" ]; then
      ready=false
      break
    fi
    state="$(docker inspect -f '{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{end}}' "$container_id")"
    case "$state" in
      "running healthy"|"running ") ;;
      *) ready=false; break ;;
    esac
  done

  if [ "$ready" = true ]; then
    echo "Deployment completed successfully."
    compose ps
    exit 0
  fi

  sleep 3
  elapsed=$((elapsed + 3))
done

echo "Deployment did not become healthy within ${timeout_seconds}s." >&2
compose ps >&2
compose logs --tail=100 api bot worker nginx >&2
exit 1
