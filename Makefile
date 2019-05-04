# Initialize sources of potential semantic changes.
SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)

# Define high-level build targets.
.PHONY: all
all: darwin linux windows test lint

.PHONY: linux
linux: bin/linux/baronial.gz

.PHONY: darwin
darwin: bin/darwin/baronial.gz

.PHONY: windows
windows: bin/windows/baronial.exe

.PHONY: docker
docker: bin/docker/baronial-alpine.tar.gz bin/docker/baronial-debian.tar.gz bin/docker/baronial-fedora.tar.gz

version.txt: ${SRC}
	sh ./get-version.sh > version.txt

revision.txt: ${SRC}
	sh ./get-revision.sh > revision.txt

# Define specific build targets.
bin/darwin/baronial: ${SRC} go.sum version.txt
	mkdir -p bin/darwin
	GOOS=darwin go build -ldflags "-X github.com/marstr/baronial/cmd.revision=$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$(cat ./version.txt)" -o bin/darwin/baronial

bin/darwin/baronial.gz: bin/darwin/baronial
	gzip -kf bin/darwin/baronial

bin/linux/baronial: ${SRC} go.sum version.txt
	mkdir -p bin/linux
	GOOS=linux go build -ldflags "-X github.com/marstr/baronial/cmd.revision=$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$(cat ./version.txt))" -o bin/linux/baronial

bin/linux/baronial.gz: bin/linux/baronial
	gzip -kf bin/linux/baronial

bin/linux/baronial-%.rpm: ${SRC} go.sum version.txt packaging/redhat/baronial.spec packaging/redhat/redhatify-version.pl
	docker build -t mars

bin/windows/baronial.exe: ${SRC} go.sum
	mkdir -p bin/windows
	GOOS=windows go build -ldflags "-X github.com/marstr/baronial/cmd.revision=$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$(cat ./version.txt)" -o bin/windows/baronial.exe

bin/docker/baronial-alpine.tar.gz: ${SRC} Dockerfile.alpine
	mkdir -p bin/docker
	docker build -t marstr/baronial:alpine -f Dockerfile.alpine .
	docker save marstr/baronial:alpine | gzip > bin/docker/baronial-alpine.tar.gz

bin/docker/baronial-debian.tar.gz: ${SRC} Dockerfile.debian
	mkdir -p bin/docker
	docker build -t marstr/baronial:debian -f Dockerfile.debian .
	docker save marstr/baronial:debian | gzip > bin/docker/baronial-debian.tar.gz

bin/docker/baronal-fedora.tar.gz: ${SRC} Dockerfile.fedora
	mkdir -p bin/docker
	docker build -t marstr/baronial:fedora -f Dockerfile.fedora .
	docker save marstr/baronial:fedora | gzip > bin/docker/baronial-fedora.tar.gz

baronial.tar.gz: ${SRC} LICENSE version.txt revision.txt
	bash ./archive-src.sh

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
install: ${SRC} version.txt
	go install -ldflags "-X github.com/marstr/baronial/cmd.revision=$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$(cat ./version.txt)"

# Remove all build products from the current system.
.PHONY: clean
clean:
	rm -rf bin
	rm -f revision.txt version.txt baronial.tar.gz
	docker rmi -f marstr/baronial:debian ||	docker rmi -f marstr/baronial:alpine || docker rmi -f marstr/baronial:fedora
