#!/usr/bin/env bash
read -p "username: " username
read -p "password: " password

# if running in docker deployment
if [[ -d ~/db ]]
then
	sqlite3 ~/db/sqlite-database.db "INSERT INTO userCredentials(username, password) VALUES('$username', '$password');"
else
  # this might need to be deleted
	sqlite3 ./../sqlite-database.db "INSERT INTO userCredentials(username, password) VALUES('$username', '$password');"
fi
