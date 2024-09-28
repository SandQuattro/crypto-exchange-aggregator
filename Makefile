.PHONY: dc lint

dc:
	docker-compose up --remove-orphans --build

lint:
	gofumpt -w ./..
	golangci-lint run --fix