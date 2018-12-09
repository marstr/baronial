#! /bin/bash

export revision="$(git rev-parse HEAD)"

if ! [[ -z "$(git status --short)" ]]; then
	export revision="${revision}-modified"
fi

echo "${revision}"
