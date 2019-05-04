#!/usr/bin/env bash

set -ev

dest="baronial-$(cat ./version.txt)"
mkdir -p ${dest}
cp -r -t ${dest} cmd/ internal/ go.mod go.sum Makefile LICENSE version.txt revision.txt main.go
tar -czvf "baronial.tar.gz" ${dest}
rm -rf ${dest}
