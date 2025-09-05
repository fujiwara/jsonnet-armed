.PHONY: clean test

jsonnet-armed: go.* *.go
	go build -o $@ ./cmd/jsonnet-armed

clean:
	rm -rf jsonnet-armed dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/jsonnet-armed/cmd/jsonnet-armed

dist:
	goreleaser build --snapshot --clean
