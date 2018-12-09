SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)

ledger: ${SRC} .git/HEAD go.sum
	go build -ldflags "-X github.com/marstr/ledger/cmd.revision=$(shell git rev-parse HEAD)"

go.sum: go.mod
	go mod download

.PHONY: test
test: ${SRC} ${TEST_SRC} .git/HEAD
	go test -v ./...

.PHONY: install
install: ledger
	cp ./ledger ${GOPATH}/bin

.PHONY: clean
clean:
	rm ledger