.PHONY: dc lint

dc:
	docker-compose up --remove-orphans --build

lint:
	golangci-lint run