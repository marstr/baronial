#!/usr/bin/env bash

version=$(./ancestors.sh | ./max-version.pl)
revision=$(./get-revision.sh)


if [[ $(git rev-parse ${version}) != ${revision} ]]; then
    version="${version}-modified"
fi

echo ${version}
