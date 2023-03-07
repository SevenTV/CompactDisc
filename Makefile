.PHONY: build lint deps dev_deps clean work dev

BUILDER := "unknown"
VERSION := "unknown"

ifeq ($(origin CD_BUILDER),undefined)
	BUILDER = $(shell git config --get user.name);
else
	BUILDER = ${CD_BUILDER};
endif

ifeq ($(origin CD_VERSION),undefined)
	VERSION = $(shell git rev-parse HEAD);
else
	VERSION = ${CD_VERSION};
endif

build:
	GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o out/compactdisc cmd/*.go

lint:
	golangci-lint run --go=1.18
	yarn prettier --check .

format:
	gofmt -s -w .
	yarn prettier --write .

deps:
	go mod download

dev_deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	yarn

test:
	go test -count=1 -cover -parallel $$(nproc) -race ./...

clean:
	rm -rf \
		out \
		node_modules

work:
	echo -e "go 1.18\n\nuse (\n\t.\n\t../Common\n\t../API\n)" > go.work
	go mod tidy

dev:
	go run cmd/main.go
