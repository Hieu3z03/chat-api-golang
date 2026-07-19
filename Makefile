# Import .env file
ifneq (,$(wildcard ./.env))
		include .env
		export $(shell sed 's/=.*//' .env)
endif

# Commands
dep: 
	go mod tidy

run: 
	go run cmd/main.go

build: 
	go build -o main cmd/main.go

run-build: build
	./main

test:
	go test -v ./...

test-auth:
	go test -v ./modules/auth/tests/...

test-user:
	go test -v ./modules/user/tests/...

test-all:
	go test -v ./modules/.../tests/...

test-coverage:
	go test -v -coverprofile=coverage.out ./modules/.../tests/...
	go tool cover -html=coverage.out

module:
	@if [ -z "$(name)" ]; then echo "Usage: make module name=<module_name>"; exit 1; fi
	@./create_module.sh $(name)

# Database commands
migrate:
	go run cmd/main.go --migrate:run

migrate-rollback:
	go run cmd/main.go --migrate:rollback

migrate-rollback-batch:
	@if [ -z "$(batch)" ]; then echo "Usage: make migrate-rollback-batch batch=<batch_number>"; exit 1; fi
	go run cmd/main.go --migrate:rollback $(batch)

migrate-rollback-all:
	go run cmd/main.go --migrate:rollback:all

migrate-status:
	go run cmd/main.go --migrate:status

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	go run cmd/main.go --migrate:create:$(name)

seed: 
	go run cmd/main.go --seed

migrate-seed: 
	go run cmd/main.go --migrate:run --seed
