#!/bin/bash

set -e

update-license-header() {
    perl -i -0pe 's/\/\/ Copyright .* Stratumn.*\n(\/\/.*\n)*/`cat LICENSE_HEADER`/ge' $1
}

extensions="go proto md"

for e in $extensions; do
	for f in $(find . -regex ".*\.$e" | grep -v vendor); do
        echo Updating $f...
        update-license-header $f
	done
done
