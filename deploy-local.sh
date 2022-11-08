#!/bin/sh
docker context use default
docker-compose -f local.yml down
docker-compose build
docker-compose -f local.yml up -d
