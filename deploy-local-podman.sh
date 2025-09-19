#!/bin/sh
podman rm -f "$IMAGE_NAME" 2>/dev/null || true
podman build -t "$IMAGE_NAME" .
podman run -p 8080:80 \
  -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
  -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
  -e AWS_REGION=us-east-2 \
  andrewwillette.com

IMAGE_NAME="andrewwillette.com"



