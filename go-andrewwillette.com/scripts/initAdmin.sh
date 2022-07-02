#!/usr/bin/env bash
read -p "username: " username
read -p "password: " password

# if running in docker deployment
if [[ -d /goApp/db ]]
then
	sqlite3 /goApp/db/sqlite-database.db "INSERT INTO userCredentials(username, password) VALUES('$username', '$password');"
else
	sqlite3 ./../sqlite-database.db "INSERT INTO userCredentials(username, password) VALUES('$username', '$password');"
fi
