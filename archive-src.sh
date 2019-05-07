#!/usr/bin/env bash

###############################################################################
# This script copies the minimal set of files required for a source build of  #
# this application into a compressed archive file. This file will be smaller  #
# than the one produced automatically by GitHub, making it a better candidate #
# for distribution to packagers.                                              #
#                                                                             #
# Input: None                                                                 #
# Output: None                                                                #
###############################################################################

set -e

dest="baronial-$(cat ./version.txt)"
mkdir -p ${dest}
cp -r -t ${dest} cmd/ internal/ go.mod go.sum Makefile LICENSE version.txt revision.txt main.go
tar -czf "baronial.tar.gz" ${dest}
rm -rf ${dest}
