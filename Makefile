build:
	go build -o bin/kube-ecr-tagger

test:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out
