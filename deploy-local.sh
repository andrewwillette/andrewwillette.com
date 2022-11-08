#!/bin/sh
export GIT_COMMIT=$(git rev-parse HEAD)
export ENV="local"
docker context use default
docker-compose down
docker-compose build
docker-compose -f local.yml up -d
