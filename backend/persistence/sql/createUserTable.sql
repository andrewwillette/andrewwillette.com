CREATE TABLE userCredentials
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "username"    TEXT    NOT NULL,
    "password"    TEXT    NOT NULL,
    "bearerToken" BLOB
)
