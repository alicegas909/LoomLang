.PHONY: build run test race vet fmt fuzz tidy spike clean

BIN := bin/loom

build:
	go build -o $(BIN) ./cmd/loom

run:
	go run ./cmd/loom

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

fuzz:
	go test ./internal/parser -fuzz=FuzzParse -fuzztime=30s

tidy:
	go mod tidy

spike:
	cd spikes/audiogc && go run .

clean:
	rm -rf bin
