.DEFAULT_GOAL := help

.PHONY: help
help: ## Help
	@echo How to use make
	@echo make test ##### run all tests in the project
	@echo make run  ##### build and run the app using docker compose, you can stop it by doing make stop
	@echo make stop ##### stop the app using docker compose

.PHONY: test
test: ## Run Tests into the packages
	@echo "Running tests"
	go mod tidy
	go test -v -covermode=atomic -coverpkg=./... ./...

.PHONY: run
run: ## Build the binary
	@echo "build and run"
	docker compose up --build -d

.PHONY: stop
stop: ## Build the binary
	@echo "stop"
	docker compose down
