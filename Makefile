.PHONY: dc lint

dc: lint
	docker-compose up --remove-orphans --build

lint:
	gofumpt -w ./..
	golangci-lint run --fix