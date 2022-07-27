#!/bin/sh
export GIT_COMMIT=$(git rev-parse HEAD)
docker context use webBeta
docker-compose build
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
