#!/bin/sh
docker context use default
docker-compose -f docker-compose-local.yml down
docker-compose -f docker-compose-local.yml build
docker-compose -f docker-compose-local.yml up -d
