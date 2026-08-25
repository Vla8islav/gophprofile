.PHONY: generate test

generate:
	go generate ./internal/mocks

test:
	go test ./...

swagger:
	swag init -g cmd/gophprofile-server/main.go -o docs --parseInternal --parseDependency