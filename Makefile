# Initialize sources of potential semantic changes.
SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)

# Define the current git revision being packed into each of the build products.
REVISION = $(shell sh ./get-revision.sh)

# Define high-level build targets.
.PHONY: all
all: darwin linux windows docker

.PHONY: linux
linux: bin/linux/baronial.gz

.PHONY: darwin
darwin: bin/darwin/baronial.gz

.PHONY: windows
windows: bin/windows/baronial.exe

.PHONY: docker
docker: bin/docker/baronial-alpine.tar.gz bin/docker/baronial-debian.tar.gz

# Define specific build targets.
bin/darwin/baronial: ${SRC} .git/HEAD go.sum
	GOOS=darwin go build -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}" -o bin/darwin/baronial

bin/darwin/baronial.gz: bin/darwin/baronial
	gzip -kf bin/darwin/baronial

bin/linux/baronial: ${SRC} .git/HEAD go.sum
	GOOS=linux go build -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}" -o bin/linux/baronial

bin/linux/baronial.gz: bin/linux/baronial
	gzip -kf bin/linux/baronial

bin/windows/baronial.exe: ${SRC} .git/HEAD go.sum
	GOOS=windows go build -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}" -o bin/windows/baronial.exe

bin/docker/baronial-alpine.tar.gz: ${SRC} Dockerfile.alpine
	mkdir -p bin/docker
	docker build -t marstr/baronial:alpine -f Dockerfile.alpine .
	docker save marstr/baronial:alpine | gzip > bin/docker/baronial-alpine.tar.gz

bin/docker/baronial-debian.tar.gz: ${SRC} Dockerfile.debian
	mkdir -p bin/docker
	docker build -t marstr/baronial:debian -f Dockerfile.debian .
	docker save marstr/baronial:debian | gzip > bin/docker/baronial-debian.tar.gz

# Ensure that the Go dependency tree is satisfied.
go.sum: go.mod
	go mod verify

go.mod: ${SRC}
	go mod tidy

# Run tests and linters
.PHONY: test
test: ${SRC} ${TEST_SRC} .git/HEAD
	go test ./...

lint: ${SRC} ${TEST_SRC}
	go vet ./...
	golint -set_exit_status ./...

# Install this build on the local system.
.PHONY: install
install: ${SRC}
	go install -ldflags "-X github.com/marstr/baronial/cmd.revision=${REVISION}"

# Remove all build products from the current system.
.PHONY: clean
clean:
	rm -rf bin
	docker rmi -f marstr/baronial:debian
	docker rmi -f marstr/baronial:alpine