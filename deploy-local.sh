#!/bin/sh
export GIT_COMMIT=$(git rev-parse HEAD)
cd backend
env GOOS=linux GOARCH=arm64 go build .
cd ..
docker context use default
docker-compose build
docker-compose -f docker-compose.yml up -d
