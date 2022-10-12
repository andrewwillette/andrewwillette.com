#!/bin/sh
read -p "ip address: " ipAddress
read -p "username: " usernameVal
read -p "password: " passwordVal

toExecute="sqlite3 /Users/andrewwillette/db/sqlite-database.db \"INSERT INTO userCredentials(username, password) VALUES('\"$usernameVal\"', '\"$passwordVal\"');\""
ssh $ipAddress $toExecute
