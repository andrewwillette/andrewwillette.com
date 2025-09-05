#!/bin/sh
docker context use desktop-linux
docker compose -f docker-compose-local.yml down
docker compose -f docker-compose-local.yml build
docker compose -f docker-compose-local.yml up -d
