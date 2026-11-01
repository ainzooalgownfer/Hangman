.PHONY: help run build test vet fmt check race load docker stop clean

BIN := bin/hangman

help: ## List the targets
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-8s %s\n", $$1, $$2}'

run: ## Start the server
	go run ./cmd/server

build: ## Build the server into bin/
	go build -o $(BIN) ./cmd/server

test: ## Run the tests
	go test ./...

vet: ## Catch suspicious code
	go vet ./...

fmt: ## Format the code
	gofmt -w .

check: fmt vet test ## Format, vet and test: run before every commit

race: ## Tests with the race detector 
	CGO_ENABLED=1 go test -race ./...

load: ## 200 players at once 
	bash scripts/simulate.sh 200

docker: ## Build the Docker image
	docker build -t hangman .

stop: ## Stop the server running in Docker (docker compose up)
	docker compose down

clean: ## Delete built binaries
	rm -rf bin
