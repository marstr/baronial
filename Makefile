SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)

REVISION = $(shell ./get-revision.sh)

.PHONY: all
all: bin/darwin/baronial.gz bin/linux/baronial.gz bin/windows/baronial.exe

bin/darwin/baronial: ${SRC} .git/HEAD go.sum
	GOOS=darwin go build -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}" -o bin/darwin/baronial

bin/darwin/baronial.gz: bin/darwin/baronial
	gzip -k bin/darwin/baronial

bin/linux/baronial: ${SRC} .git/HEAD go.sum
	GOOS=linux go build -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}" -o bin/linux/baronial

bin/linux/baronial.gz: bin/linux/baronial
	gzip -k bin/linux/baronial

bin/windows/baronial.exe: ${SRC} .git/HEAD go.sum
	GOOS=windows go build -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}" -o bin/windows/baronial.exe

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
	rm -rf bin