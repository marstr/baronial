SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)

REVISION = $(shell ./get-revision.sh)

baronial: ${SRC} .git/HEAD go.sum
	go build -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}"

go.sum: go.mod
	go mod verify

go.mod: ${SRC}
	go mod tidy

.PHONY: test
test: ${SRC} ${TEST_SRC} .git/HEAD
	go test -v ./...
	go vet ./...
	golint -set_exit_status ./...

.PHONY: install
install: baronial
	cp ./baronial ${GOPATH}/bin

.PHONY: clean
clean:
	rm baronial