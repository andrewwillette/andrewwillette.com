#!/bin/sh
export GIT_COMMIT=$(git rev-parse HEAD)
docker context use default
docker-compose build
make deploy-local
docker-compose -f docker-compose.yml up
