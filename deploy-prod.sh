#!/bin/sh
export GIT_COMMIT=$(git rev-parse HEAD)
docker context use andrewwilletteWebsiteAws
docker-compose build
docker-compose -f docker-compose.yml up -d
