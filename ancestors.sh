#!/usr/bin/env bash

for tag in $(git tag) ; do
    if git merge-base --is-ancestor ${tag} HEAD; then
        echo ${tag}
    fi
done