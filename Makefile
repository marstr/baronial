# Initialize sources of potential semantic changes.
SRC = $(shell find . -name '*.go' -type f)
TEST_SRC = $(shell find . -name '*_test.go' -type f)
DOCKER?=docker

# Define high-level build targets.
.PHONY: all
all: darwin linux windows test lint docker rpm

.PHONY: linux
linux: bin/linux/baronial.gz

.PHONY: darwin
darwin: bin/darwin/baronial.gz

.PHONY: windows
windows: bin/windows/baronial.exe

.PHONY: docker
docker: bin/docker/baronial-alpine.tar.gz bin/docker/baronial-debian.tar.gz bin/docker/baronial-fedora33.tar.gz bin/docker/baronial-fedora34.tar.gz bin/docker/baronial-el8.tar.gz

.PHONY: fedora
fedora: fedora33 fedora34 fedora35

.PHONY: fedora33
fedora33: bin/linux/baronial.fc33.src.rpm bin/linux/baronial.fc33.x86_64.rpm bin/docker/baronial-fedora33.tar.gz

.PHONY: fedora34
fedora34: bin/linux/baronial.fc34.src.rpm bin/linux/baronial.fc34.x86_64.rpm bin/docker/baronial-fedora34.tar.gz

.PHONY: fedora35
fedora35: bin/linux/baronial.fc35.src.rpm bin/linux/baronial.fc35.x86_64.rpm bin/docker/baronial-fedora35.tar.gz

.PHONY: el8
el8: bin/linux/baronial.el8.src.rpm bin/linux/baronial.el8.x86_64.rpm bin/docker/baronial-el8.tar.gz

.PHONY: opensuse
opensuse: bin/linux/baronial.lp153.src.rpm bin/linux/baronial.lp153.x86_64.rpm bin/docker/baronial-opensuse_leap153.tar.gz

.PHONY: alpine
alpine: bin/docker/baronial-alpine.tar.gz

.PHONY: rpm
rpm: bin/linux/baronial.fc33.src.rpm bin/linux/baronial.fc34.x86_64.rpm bin/linux/baronial.fc33.src.rpm bin/linux/baronial.fc34.x86_64.rpm bin/linux/baronial.lp151.src.rpm bin/linux/baronial.lp151.x86_64.rpm bin/linux/baronial.el8.x86_64.rpm bin/linux/baronial.el8.src.rpm

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

bin/docker/baronial-fedora33.tar.gz: ${SRC} Dockerfile.fedora
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=33 -t marstr/baronial:fedora33-rpm-builder -f Dockerfile.fedora --target rpm-builder .
	${DOCKER} build --build-arg tag=33 -t marstr/baronial:fedora33 -f Dockerfile.fedora .
	${DOCKER} save marstr/baronial:fedora31 | gzip > bin/docker/baronial-fedora33.tar.gz

bin/linux/baronial.fc33.src.rpm: bin/docker/baronial-fedora33.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora33-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc33.src.rpm > bin/linux/baronial.fc33.src.rpm

bin/linux/baronial.fc33.x86_64.rpm: bin/docker/baronial-fedora33.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora33-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc33.x86_64.rpm > bin/linux/baronial.fc33.x86_64.rpm

bin/docker/baronial-fedora34.tar.gz: ${SRC} Dockerfile.fedora
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=34 -t marstr/baronial:fedora34-rpm-builder -f Dockerfile.fedora --target rpm-builder .
	${DOCKER} build --build-arg tag=34 -t marstr/baronial:fedora34 -f Dockerfile.fedora .
	${DOCKER} save marstr/baronial:fedora34 | gzip > bin/docker/baronial-fedora34.tar.gz

bin/linux/baronial.fc34.src.rpm: bin/docker/baronial-fedora34.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora34-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc34.src.rpm > bin/linux/baronial.fc34.src.rpm

bin/linux/baronial.fc34.x86_64.rpm: bin/docker/baronial-fedora34.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora34-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc34.x86_64.rpm > bin/linux/baronial.fc34.x86_64.rpm

bin/docker/baronial-fedora35.tar.gz: ${SRC} Dockerfile.fedora
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=35 -t marstr/baronial:fedora35-rpm-builder -f Dockerfile.fedora --target rpm-builder .
	${DOCKER} build --build-arg tag=35 -t marstr/baronial:fedora35 -f Dockerfile.fedora .
	${DOCKER} save marstr/baronial:fedora35 | gzip > bin/docker/baronial-fedora35.tar.gz

bin/linux/baronial.fc35.src.rpm: bin/docker/baronial-fedora35.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora35-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc35.src.rpm > bin/linux/baronial.fc35.src.rpm

bin/linux/baronial.fc35.x86_64.rpm: bin/docker/baronial-fedora35.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:fedora35-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.fc35.x86_64.rpm > bin/linux/baronial.fc35.x86_64.rpm

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

bin/docker/baronial-opensuse_leap153.tar.gz: ${SRC} Dockerfile.opensuse_leap
	mkdir -p bin/docker
	${DOCKER} build --build-arg tag=15.3 -t marstr/baronial:leap153-rpm-builder -f Dockerfile.opensuse_leap --target rpm-builder .
	${DOCKER} build --build-arg tag=15.3 -t marstr/baronial:leap153 -f Dockerfile.opensuse_leap --target rpm-builder .
	${DOCKER} save marstr/baronial:leap153 | gzip > bin/docker/baronial-opensuse_leap153.tar.gz

bin/linux/baronial.lp153.src.rpm: bin/docker/baronial-opensuse_leap153.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:leap153-rpm-builder cat /root/rpmbuild/SRPMS/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.lp153.src.rpm > bin/linux/baronial.lp153.src.rpm

bin/linux/baronial.lp153.x86_64.rpm: bin/docker/baronial-opensuse_leap153.tar.gz version.txt
	mkdir -p bin/linux
	${DOCKER} run --rm marstr/baronial:leap153-rpm-builder cat /root/rpmbuild/RPMS/x86_64/baronial-$$(cat ./version.txt | ./packaging/redhat/redhatify-version.pl)-1.lp153.x86_64.rpm > bin/linux/baronial.lp153.x86_64.rpm

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
	${DOCKER} rmi -f marstr/baronial:fedora33-rpm-builder 2>/dev/null || echo 'Skipping Fedora 33 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora33 2>/dev/null || echo 'Skipping Fedora 33 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora34-rpm-builder 2>/dev/null || echo 'Skipping Fedora 34 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora34 2>/dev/null || echo 'Skipping Fedora 34 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora35-rpm-builder 2>/dev/null || echo 'Skipping Fedora 35 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:fedora35 2>/dev/null || echo 'Skipping Fedora 35 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:leap151-rpm-builder 2>/dev/null || echo 'Skipping openSUSE Leap 15.1 RPM Builder Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:leap151 2>/dev/null || echo 'Skipping openSUSE Leap 15.1 Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:el8-rpm-builder 2>/dev/null || echo 'Skipping Enterprise Linux 8 RPM Build Docker Image Delete' > /dev/stderr
	${DOCKER} rmi -f marstr/baronial:el8 2>/dev/null || echo 'Skipping Enterprise Linux 8 Docker Image Delete' > /dev/null

