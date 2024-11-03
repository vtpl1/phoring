all: build

lintall: lintverify lint

lint: clear lint1

# Build the application
build:
	@echo "Building..."
	@go build

# Run the application
run:
	@go run .

# Run the application
lintverify:
	@/home/vscode/go/bin/golangci-lint config verify

# Run the application
lint1:
	@/home/vscode/go/bin/golangci-lint run ...

# Run the application
vul:
	@/home/vscode/go/bin/govulncheck -show verbose ./...

clear:
	@clear && printf '\e[3J'

.PHONY: all build run test clean lintverify lint lintall