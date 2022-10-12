#!/usr/bin/env bash

starting_dir=$(pwd)

# build go executable in production version and start serving on port 6969
cd ../backend || return
go build -ldflags "-s -w"
nohup ./willette_api &
cd "$starting_dir" || return

# build react executable and start serving (using nginx) on port 80
cd ../frontend || return
npm install
npm run build
nohup npm run start-prod &
