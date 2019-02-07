#! /bin/bash

revision="$(git rev-parse HEAD)"

if ! [[ -z "$(git status --short)" ]]; then
	revision="${revision}-modified"
fi

echo "${revision}"
