VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

.PHONY: build-cli build-server build-all clean test

build-cli:
	go build -ldflags "$(LDFLAGS)" -o bin/agentskills .

build-server:
	go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server .

build-all: build-all-cli build-all-server

build-all-cli:
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-windows-amd64.exe .

build-all-server:
	GOOS=linux   GOARCH=amd64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-windows-amd64.exe .

test:
	go test -v -race -tags server ./...

clean:
	rm -rf bin/
