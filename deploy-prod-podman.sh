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

podman build -f Dockerfile.prod -t "$IMAGE_NAME" .
podman save "$IMAGE_NAME" -o "$TAR_FILE"

ssh "$EC2_USER@$EC2_HOST" "mkdir -p $REMOTE_DIR"
scp "$TAR_FILE" "$EC2_USER@$EC2_HOST:$REMOTE_DIR/"

rm -f "$TAR_FILE"

ssh "$EC2_USER@$EC2_HOST" <<EOF
  set -euo pipefail
  cd "$REMOTE_DIR"

  sudo podman load -i "$TAR_FILE"

  rm -f "$TAR_FILE"

  echo "Stopping and removing previous container if exists..."
  sudo podman rm -f "$IMAGE_NAME" 2>/dev/null || true

  sudo podman container prune -f
  sudo podman image prune -f
  sudo podman volume prune -f

  echo "Running new container..."
  # /app/logs defined in Dockerfile, server.go configs cache dir on server
  sudo podman run -d --name "$IMAGE_NAME" \
    -p 80:80 -p 443:443 \
    -v "$CACHE_DIR:/var/www/.cache" \
    -v "$LOG_DIR:/app/logs" \
    "localhost/$IMAGE_NAME:latest"
EOF

echo "✅ Deployment complete! Visit http://$EC2_HOST"
