build:
	@go build -o bin/goadopt

run: build
	@./bin/goadopt

test:
	@go test -v ./...