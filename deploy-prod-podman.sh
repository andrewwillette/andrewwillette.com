#!/bin/bash
set -euo pipefail

# Set your EC2 instance IP and user
EC2_USER=ubuntu
EC2_HOST="$ANDREWWILLETTE_PUBLIC_IP"
REMOTE_DIR="/home/ubuntu/andrewwillette.com"
IMAGE_NAME="andrewwillette.com"

# 1. Build container image locally
podman build -t "$IMAGE_NAME" .

# 2. Save container image to a tarball
podman save "$IMAGE_NAME" -o "$IMAGE_NAME.tar"

ssh "$EC2_USER@$EC2_HOST" "mkdir -p $REMOTE_DIR"
# 3. Copy tarball and project files to EC2 instance
scp "$IMAGE_NAME.tar" "$EC2_USER@$EC2_HOST:$REMOTE_DIR/"

# 4. SSH into EC2 and load the image, then run it
ssh "$EC2_USER@$EC2_HOST" <<EOF
  set -e
  cd "$REMOTE_DIR"
  sudo podman load -i "$IMAGE_NAME.tar"

  # Stop any previous container
  podman rm -f "$IMAGE_NAME" 2>/dev/null || true

  # Run the container on port 80 (requires sudo/root or rootless port binding support)
  sudo podman run -d --name "$IMAGE_NAME" \
    -p 80:80 \
    "localhost/$IMAGE_NAME:latest"
EOF
