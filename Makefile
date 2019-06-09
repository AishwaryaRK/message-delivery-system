PKGS = $(shell go list ./... | grep -v /test)

build:
    CGO_ENABLED=0 go build ./...
.PHONY: build

lint:
	golint $(PKGS) 
.PHONY: lint

test-unit: 
	go test --race --cover -v $(PKGS)
.PHONY: test-unit

test-integration:
	go test --race -v test/integration_test.go
.PHONY: test-integration

test-benchmark:
	go test -v -bench=. test/benchmark_test.go
.PHONY: test-benchmark

test: test-unit test-integration
.PHONY: test
