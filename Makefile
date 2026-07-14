.PHONY: generate test

generate:
	./scripts/generate-proto.sh

test: generate
	go test ./...
