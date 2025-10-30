# Root Makefile - engine library only

.PHONY: test test-all tidy

test:
	go test ./pkg/...

test-all:
	go test ./...

tidy:
	go mod tidy
