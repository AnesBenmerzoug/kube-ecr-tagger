.PHONY= all test build

BINARY_DIRECTORY=bin
BINARY_NAME=$(BINARY_DIRECTORY)/kube-ecr-tagger

all: test build

build: $(BINARY_NAME)

$(BINARY_NAME):
	go build -o $(BINARY_NAME)

build-static:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o $(BINARY_NAME)

build-image:
	docker build . -t kube-ecr-tagger:latest

test:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out

format:
	gofmt -l -w -s .

lint:
	golangci-lint run

install-golangci-lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(HOME)/bin v1.22.0

lint-ci:
	$(HOME)/bin/golangci-lint run

clean:
	go clean
	rm -rf bin
