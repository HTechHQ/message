.PHONE:help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'




.PHONY:dev-setup
dev-setup: ## Initialise this machine with development dependencies
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.49.0




.PHONY:static-check
static-check: ## Run static code checks
	golangci-lint run

.PHONY: test-unit
test-unit:
	go test -race ./...

.PHONY: test
test: static-check ## Run all tests without any cache
	go test -race ./... -count=1 -coverprofile cover.out
	go tool cover -func cover.out | grep total:




.PHONY: show-cover
show-cover:
	go tool cover -html=cover.out -o cover.html
	xdg-open cover.html
