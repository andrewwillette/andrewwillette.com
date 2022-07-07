deploy-local:
	docker context use default
	docker-compose -f docker-compose.yml up
deploy-prod:
	docker context use webBeta
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
