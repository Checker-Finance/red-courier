APP_NAME = red-courier
CMD_PATH = ./cmd/courier
BINARY = $(APP_NAME)
VERSION ?= dev

# Go flags
GO_FILES := $(shell find . -name '*.go' -not -path "./vendor/*")
LD_FLAGS = -X main.version=$(VERSION)

.PHONY: all build test clean run docker docker-run

all: build

build:
	go build -ldflags="$(LD_FLAGS)" -o $(BINARY) $(CMD_PATH)

test:
	go test ./...

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

docker:
	docker build -t $(APP_NAME):$(VERSION) .

docker-run: docker
	docker run --rm -v $$PWD/config.yaml:/red-courier/config.yaml $(APP_NAME):$(VERSION)