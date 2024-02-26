build:
	sam build

run: build
	sam local start-api

