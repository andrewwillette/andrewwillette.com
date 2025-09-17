#!/bin/sh
podman build -t andrewwillette.com .
podman run -p 8080:80 \
  -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
  -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
  -e AWS_REGION=us-east-2 \
  andrewwillette.com
