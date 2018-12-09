SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)

baronial: ${SRC} .git/HEAD go.sum
	go build -ldflags "-X github.com/marstr/baronial/cmd.revision=$(shell git rev-parse HEAD)"

go.sum: go.mod
	go mod download

.PHONY: test
test: ${SRC} ${TEST_SRC} .git/HEAD
	go test -v ./...

.PHONY: install
install: baronial
	cp ./baronial ${GOPATH}/bin

.PHONY: clean
clean:
	rm baronial