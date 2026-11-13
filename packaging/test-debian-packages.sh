#!/usr/bin/env bash

set -euo pipefail

packages=$(realpath "${1:-bin/debian}")
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

test "$(dpkg-deb -f "$packages"/baronial_*.deb Package)" = baronial
dpkg-deb --extract "$packages"/baronial_*.deb "$work/binary"
"$work/binary/usr/bin/baronial" --help > /dev/null
test -s "$work/binary/usr/share/doc/baronial/copyright"
test -d "$work/binary/usr/share/doc/baronial/vendor"

dpkg-source --extract "$packages"/baronial_*.dsc "$work/source"
test -s "$work/source/vendor/modules.txt"
test -s "$work/source/version.txt"
test -s "$work/source/revision.txt"
version=$(sed -e 's/^[vV]//' -e 's/-/~/g' "$work/source/version.txt")
test "$(dpkg-deb -f "$packages"/baronial_*.deb Version)" = "$version-1"
"$work/binary/usr/bin/baronial" version | grep -F -- "$(cat "$work/source/revision.txt")"
cd "$work/source"
GOPROXY=off GOSUMDB=off GOMODCACHE="$work/go-mod" GOCACHE="$work/go-build" \
    dpkg-buildpackage --build=binary -us -uc
