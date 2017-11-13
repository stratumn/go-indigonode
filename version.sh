#!/bin/sh

set -e

version_file=`dirname $0`/VERSION

if [ -f $version_file ]; then
    echo v$(cat $version_file)
    exit 0
fi

set +e
current_tag=`git describe --tags 2> /dev/null`
set -e

if git tag -l | grep "^$current_tag\$"; then
    echo $current_tag
else
    git branch | grep '^\*' | awk '{ print $2}'
fi
