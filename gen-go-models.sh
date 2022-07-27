#!/bin/sh
oapi-codegen -package=genmodels backend/openapi.yaml > backend/server/models/models.go
