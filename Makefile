.PHONY: clean test

jsonnet-armed: go.* *.go cmd/*/*.go functions/*.go
	go build -o $@ ./cmd/jsonnet-armed

clean:
	rm -rf jsonnet-armed dist/

test:
	go test -race -v ./...

install:
	go install github.com/fujiwara/jsonnet-armed/cmd/jsonnet-armed

dist:
	goreleaser build --snapshot --clean
