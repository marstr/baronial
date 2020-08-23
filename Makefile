# Initialize sources of potential semantic changes.
SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)
DOCKER?=docker

# Define high-level build targets.
.PHONY: all
all: darwin linux windows test lint baronial.tar.gz

.PHONY: linux
linux: bin/linux/baronial.gz

.PHONY: darwin
darwin: bin/darwin/baronial.gz

.PHONY: windows
windows: bin/windows/baronial.exe

.PHONY: docker
docker: bin/docker/baronial-alpine.tar.gz bin/docker/baronial-debian.tar.gz bin/docker/baronial-fedora30.tar.gz bin/docker/baronial-fedora31.tar.gz

.PHONY: fedora
fedora: fedora30 fedora31 fedora32

.PHONY: fedora30
fedora30: bin/linux/baronial.fc30.src.rpm bin/linux/baronial.fc30.x86_64.rpm bin/docker/baronial-fedora30.tar.gz

.PHONY: fedora31
fedora31: bin/linux/baronial.fc31.src.rpm bin/linux/baronial.fc31.x86_64.rpm bin/docker/baronial-fedora31.tar.gz

.PHONY: fedora32
fedora32: bin/linux/baronial.fc32.src.rpm bin/linux/baronial.fc32.x86_64.rpm bin/docker/baronial-fedora32.tar.gz

.PHONY: el8
el8: bin/linux/baronial.el8.src.rpm bin/linux/baronial.el8.x86_64.rpm bin/docker/baronial-el8.tar.gz

.PHONY: opensuse
opensuse: bin/linux/baronial.lp151.src.rpm bin/linux/baronial.lp151.x86_64.rpm bin/docker/baronial-opensuse_leap151.tar.gz

.PHONY: alpine
alpine: bin/docker/baronial-alpine.tar.gz

.PHONY: rpm
rpm: bin/linux/baronial.fc29.src.rpm bin/linux/baronial.fc29.x86_64.rpm bin/linux/baronial.fc30.src.rpm bin/linux/baronial.fc30.x86_64.rpm bin/linux/baronial.lp151.src.rpm bin/linux/baronial.lp151.x86_64.rpm bin/linux/baronial.el8.x86_64.rpm bin/linux/baronial.el8.src.rpm

version.txt: ${SRC} go.sum
	perl ./get-version.pl > version.txt

revision.txt: ${SRC} go.sum
	perl ./get-revision.pl > revision.txt

# Define specific build targets.
bin/darwin/baronial: ${SRC} go.sum version.txt revision.txt
	mkdir -p bin/darwin
	GOOS=darwin go build -ldflags "-X github.com/marstr/baronial/cmd.revision=$$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$$(cat ./version.txt)" -o bin/darwin/baronial

bin/darwin/baronial.gz: bin/darwin/baronial
	gzip -kf bin/darwin/baronial

bin/linux/baronial: ${SRC} go.sum version.txt revision.txt
	mkdir -p bin/linux
	GOOS=linux go build -ldflags "-X github.com/marstr/baronial/cmd.revision=$$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$$(cat ./version.txt)" -o bin/linux/baronial

bin/linux/baronial.gz: bin/linux/baronial
	gzip -kf bin/linux/baronial

bin/linux/baronial-%.rpm: ${SRC} go.sum version.txt packaging/redhat/baronial.spec packaging/redhat/redhatify-version.pl
	docker build -t mars

bin/windows/baronial.exe: ${SRC} go.sum version.txt revision.txt
	mkdir -p bin/windows
	GOOS=windows go build -ldflags "-X github.com/marstr/baronial/cmd.revision=$$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$$(cat ./version.txt)" -o bin/windows/baronial.exe

bin/docker/baronial-alpine.tar.gz: ${SRC} Dockerfile.alpine
	mkdir -p bin/docker
	${DOCKER} build -t marstr/baronial:alpine -f Dockerfile.alpine .
	${DOCKER} save marstr/baronial:alpine | gzip > bin/docker/baronial-alpine.tar.gz

bin/docker/baronial-debian.tar.gz: ${SRC} Dockerfile.debian
	mkdir -p bin/docker
	${DOCKER} build -t marstr/baronial:debian -f Dockerfile.debian .
	${DOCKER} save marstr/baronial:debian | gzip > bin/docker/baronial-debian.tar.gz

bin/docker/baronial-fedora30.tar.gz: ${SRC} Dockerfile.fedora
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=30 -t marstr/baronial:fedora30-rpm-builder -f Dockerfile.fedora --target rpm-builder .
	${DOCKER} build --build-arg tag=30 -t marstr/baronial:fedora30 -f Dockerfile.fedora .
	${DOCKER} save marstr/baronial:fedora30 | gzip > bin/docker/baronial-fedora30.tar.gz

bin/linux/baronial.fc30.src.rpm: bin/docker/baronial-fedora30.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora30-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc30.src.rpm > bin/linux/baronial.fc30.src.rpm

bin/linux/baronial.fc30.x86_64.rpm: bin/docker/baronial-fedora30.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora30-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc30.x86_64.rpm > bin/linux/baronial.fc30.x86_64.rpm

bin/docker/baronial-fedora31.tar.gz: ${SRC} Dockerfile.fedora
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=31 -t marstr/baronial:fedora31-rpm-builder -f Dockerfile.fedora --target rpm-builder .
	${DOCKER} build --build-arg tag=31 -t marstr/baronial:fedora31 -f Dockerfile.fedora .
	${DOCKER} save marstr/baronial:fedora31 | gzip > bin/docker/baronial-fedora31.tar.gz

bin/linux/baronial.fc31.src.rpm: bin/docker/baronial-fedora31.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora31-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc31.src.rpm > bin/linux/baronial.fc31.src.rpm

bin/linux/baronial.fc31.x86_64.rpm: bin/docker/baronial-fedora31.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora31-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc31.x86_64.rpm > bin/linux/baronial.fc31.x86_64.rpm

bin/docker/baronial-fedora32.tar.gz: ${SRC} Dockerfile.fedora
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=32 -t marstr/baronial:fedora32-rpm-builder -f Dockerfile.fedora --target rpm-builder .
	${DOCKER} build --build-arg tag=32 -t marstr/baronial:fedora32 -f Dockerfile.fedora .
	${DOCKER} save marstr/baronial:fedora32 | gzip > bin/docker/baronial-fedora32.tar.gz

bin/linux/baronial.fc32.src.rpm: bin/docker/baronial-fedora32.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora32-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc32.src.rpm > bin/linux/baronial.fc32.src.rpm

bin/linux/baronial.fc32.x86_64.rpm: bin/docker/baronial-fedora32.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora32-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc32.x86_64.rpm > bin/linux/baronial.fc32.x86_64.rpm

bin/docker/baronial-el8.tar.gz: ${SRC} Dockerfile.rhel
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=8 -t marstr/baronial:el8-rpm-builder -f Dockerfile.rhel --target rpm-builder .
	${DOCKER} build --build-arg tag=8 -t marstr/baronial:el8 -f Dockerfile.rhel .
	${DOCKER} save marstr/baronial:el8 | gzip > bin/docker/baronial-el8.tar.gz

bin/linux/baronial.el8.src.rpm: bin/docker/baronial-el8.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:el8-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.el8.src.rpm > bin/linux/baronial.el8.src.rpm

bin/linux/baronial.el8.x86_64.rpm: bin/docker/baronial-el8.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:el8-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.el8.x86_64.rpm > bin/linux/baronial.el8.x86_64.rpm

bin/docker/baronial-opensuse_leap151.tar.gz: ${SRC} Dockerfile.opensuse_leap
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=15.1 -t marstr/baronial:leap151-rpm-builder -f Dockerfile.opensuse_leap --target rpm-builder .
	${DOCKER} build --build-arg tag=15.1 -t marstr/baronial:leap151 -f Dockerfile.opensuse_leap --target rpm-builder .
	${DOCKER} save marstr/baronial:leap151 | gzip > bin/docker/baronial-opensuse_leap151.tar.gz

bin/linux/baronial.lp151.src.rpm: bin/docker/baronial-opensuse_leap151.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:leap151-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.lp151.src.rpm > bin/linux/baronial.lp151.src.rpm

bin/linux/baronial.lp151.x86_64.rpm: bin/docker/baronial-opensuse_leap151.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:leap151-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.lp151.x86_64.rpm > bin/linux/baronial.lp151.x86_64.rpm

baronial.tar.gz: ${SRC} LICENSE version.txt revision.txt
	bash ./archive-src.sh

# Ensure that the Go dependency tree is satisfied.
go.sum: go.mod
	go mod verify

go.mod: ${SRC}
	go mod tidy

# Run tests and linters
.PHONY: test
test: .semaphores/test
.semaphores/test: ${SRC} ${TEST_SRC}
	go test ./...
	bash ./packaging/redhat/test-redhatify-version.sh
	mkdir -p .semaphores && touch .semaphores/test

report.xml:
	go test -v ./... 2>&1 | tee /dev/stderr | go-junit-report > report.xml

.PHONY: lint
lint: .semaphores/lint
.semaphores/lint: ${SRC} ${TEST_SRC}
	go vet ./...
	golint -set_exit_status ./...
	mkdir -p .semaphores && touch .semaphores/lint

# Install this build on the local system.
.PHONY: install
install: ${SRC} version.txt revision.txt
	go install -ldflags "-X github.com/marstr/baronial/cmd.revision=$$(cat ./revision.txt) -X github.com/marstr/baronial/cmd.version=$$(cat ./version.txt)"

# Remove all build products from the current system.
.PHONY: clean
clean:
	rm -rf bin
	rm -f revision.txt version.txt baronial.tar.gz
	rm -rf .semaphores
	${DOCKER} rmi -f marstr/baronial:debian 2>/dev/null || echo 'Skipping Debian Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:alpine 2>/dev/null || echo 'Skipping Alpine Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora30-rpm-builder 2>/dev/null || echo 'Skipping Fedora 30 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora30 2>/dev/null || echo 'Skipping Fedora 30 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora31-rpm-builder 2>/dev/null || echo 'Skipping Fedora 31 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora31 2>/dev/null || echo 'Skipping Fedora 31 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora32-rpm-builder 2>/dev/null || echo 'Skipping Fedora 32 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora32 2>/dev/null || echo 'Skipping Fedora 32 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:leap151-rpm-builder 2>/dev/null || echo 'Skipping openSUSE Leap 15.1 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:leap151 2>/dev/null || echo 'Skipping openSUSE Leap 15.1 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:el8-rpm-builder 2>/dev/null || echo 'Skipping Enterprise Linux 8 RPM Build Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:el8 2>/dev/null || echo 'Skipping Enterprise Linux 8 Docker Image Delete' > /dev/null

