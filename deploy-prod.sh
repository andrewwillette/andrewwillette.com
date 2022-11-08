#!/bin/sh
docker context use andrewwilletteWebsiteAws
docker-compose -f prod.yml down
docker-compose build
docker-compose -f prod.yml up -d
