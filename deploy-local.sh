#!/bin/sh
export GIT_COMMIT=$(git rev-parse HEAD)
docker context use default
docker-compose build
docker-compose -f docker-compose.yml up -d
