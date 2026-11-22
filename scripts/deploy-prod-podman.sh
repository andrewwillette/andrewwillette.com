#!/bin/bash
# Prevents running the rest of the script after a failed command & fail on any failure in a pipeline
set -euo pipefail

EC2_USER=ubuntu
EC2_HOST="$ANDREWWILLETTE_PODMAN_PUBLIC_IP"
REMOTE_DIR="/home/ubuntu/andrewwillette.com"
IMAGE_NAME="andrewwillette.com"
TAR_FILE="$IMAGE_NAME.tar"
CACHE_DIR="/var/www/.cache"
LOG_DIR="/home/ubuntu"
TRAFFIC_DB="/home/ubuntu/traffic.db"
# The website and Forgejo share a pod so the website can proxy to Forgejo
# over localhost (podman 3.x on the instance has no container-name DNS).
POD_NAME="site"
GIT_HOST="git.andrewwillette.com"
FORGEJO_NAME="forgejo"
# Bump this tag to upgrade Forgejo; the next deploy recreates its container.
FORGEJO_IMAGE="codeberg.org/forgejo/forgejo:17-rootless"
FORGEJO_DIR="/home/ubuntu/forgejo"

podman build -f Dockerfile.prod --build-arg ADMIN_PASSWORD="$PERSONAL_WEBSITE_PASSWORD" -t "$IMAGE_NAME" .
podman save "$IMAGE_NAME" -o "$TAR_FILE"

ssh "$EC2_USER@$EC2_HOST" "mkdir -p $REMOTE_DIR"
scp "$TAR_FILE" "$EC2_USER@$EC2_HOST:$REMOTE_DIR/"

rm -f "$TAR_FILE"

ssh "$EC2_USER@$EC2_HOST" \
  REMOTE_DIR="$REMOTE_DIR" IMAGE_NAME="$IMAGE_NAME" TAR_FILE="$TAR_FILE" \
  CACHE_DIR="$CACHE_DIR" LOG_DIR="$LOG_DIR" TRAFFIC_DB="$TRAFFIC_DB" \
  POD_NAME="$POD_NAME" GIT_HOST="$GIT_HOST" FORGEJO_NAME="$FORGEJO_NAME" \
  FORGEJO_IMAGE="$FORGEJO_IMAGE" FORGEJO_DIR="$FORGEJO_DIR" \
  bash -s <<'EOF'
  set -euo pipefail
  cd "$REMOTE_DIR"

  sudo podman load -i "$TAR_FILE"

  rm -f "$TAR_FILE"

  echo "Stopping and removing previous container if exists..."
  sudo podman rm -f "$IMAGE_NAME" 2>/dev/null || true

  sudo podman container prune -f
  sudo podman image prune -f
  sudo podman volume prune -f

  # Create traffic.db if it doesn't exist
  touch "$TRAFFIC_DB"

  # The pod owns the public ports; its containers share one network namespace.
  if ! sudo podman pod exists "$POD_NAME"; then
    echo "Creating pod $POD_NAME..."
    sudo podman pod create --name "$POD_NAME" -p 80:80 -p 443:443
  fi

  # Forgejo keeps running across website deploys; it is only (re)created when
  # missing or when FORGEJO_IMAGE changes. Its data lives in $FORGEJO_DIR.
  if sudo podman container exists "$FORGEJO_NAME" &&
    [ "$(sudo podman inspect "$FORGEJO_NAME" --format '{{.ImageName}}')" != "$FORGEJO_IMAGE" ]; then
    echo "Forgejo image changed, recreating container..."
    sudo podman rm -f "$FORGEJO_NAME"
  fi
  if ! sudo podman container exists "$FORGEJO_NAME"; then
    sudo podman pull "$FORGEJO_IMAGE"
    sudo mkdir -p "$FORGEJO_DIR/data"
    # The rootless image runs as uid 1000.
    sudo chown 1000:1000 "$FORGEJO_DIR/data"
    if ! sudo test -f "$FORGEJO_DIR/secrets.env"; then
      echo "Generating Forgejo secrets..."
      secret_key=$(sudo podman run --rm --entrypoint forgejo "$FORGEJO_IMAGE" generate secret SECRET_KEY)
      internal_token=$(sudo podman run --rm --entrypoint forgejo "$FORGEJO_IMAGE" generate secret INTERNAL_TOKEN)
      printf 'FORGEJO__security__SECRET_KEY=%s\nFORGEJO__security__INTERNAL_TOKEN=%s\n' \
        "$secret_key" "$internal_token" | sudo tee "$FORGEJO_DIR/secrets.env" >/dev/null
      sudo chmod 600 "$FORGEJO_DIR/secrets.env"
    fi
    echo "Running Forgejo..."
    # HTTP_ADDR=127.0.0.1 keeps Forgejo reachable only from inside the pod,
    # i.e. through the website's TLS reverse proxy.
    sudo podman run -d --name "$FORGEJO_NAME" --pod "$POD_NAME" \
      -v "$FORGEJO_DIR/data:/var/lib/gitea" \
      --env-file "$FORGEJO_DIR/secrets.env" \
      -e FORGEJO__server__DOMAIN="$GIT_HOST" \
      -e FORGEJO__server__ROOT_URL="https://$GIT_HOST/" \
      -e FORGEJO__server__HTTP_ADDR=127.0.0.1 \
      -e FORGEJO__server__HTTP_PORT=3000 \
      -e FORGEJO__server__DISABLE_SSH=true \
      -e FORGEJO__server__START_SSH_SERVER=false \
      -e FORGEJO__database__DB_TYPE=sqlite3 \
      -e FORGEJO__security__INSTALL_LOCK=true \
      -e FORGEJO__service__DISABLE_REGISTRATION=true \
      "$FORGEJO_IMAGE"
  fi

  # Starts the pod and Forgejo if they were stopped (e.g. after a reboot).
  sudo podman pod start "$POD_NAME"

  echo "Running new container..."
  # /app/logs defined in Dockerfile, server.go configs cache dir on server
  sudo podman run -d --name "$IMAGE_NAME" --pod "$POD_NAME" \
    -v "$CACHE_DIR:/var/www/.cache" \
    -v "$LOG_DIR:/app/logs" \
    -v "$TRAFFIC_DB:/app/traffic.db" \
    "localhost/$IMAGE_NAME:latest"
EOF

echo "✅ Deployment complete! Visit http://$EC2_HOST"
