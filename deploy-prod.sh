#!/bin/sh
export GIT_COMMIT=$(git rev-parse HEAD)
docker context use andrewwilletteWebsiteAws
docker-compose down
docker-compose build
docker-compose -f prod.yml up -d
