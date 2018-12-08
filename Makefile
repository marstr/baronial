SRC = $(shell find . -name '*.go' -type f)

ledger: ${SRC} .git/HEAD
	go build -ldflags "-X github.com/marstr/ledger/cmd.revision=$(shell git rev-parse HEAD)"

.PHONY: test
test: ${SRC}
	go test -v ./...

.PHONY: install
install: ledger
	cp ./ledger ${GOPATH}/bin

.PHONY: clean
clean:
	rm ledger