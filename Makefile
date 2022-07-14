deploy-local:
	docker context use default
	docker-compose -f docker-compose.yml up
deploy-prod:
	docker context use webBeta
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
gen-go-models:
	oapi-codegen -package=genmodels backend/openapi.yaml > backend/server/models/models.go
