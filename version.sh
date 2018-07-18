#!/bin/sh

version_file=`dirname $0`/VERSION

if [ -f $version_file ]; then
    echo v$(cat $version_file)
    exit 0
fi

current_tag=`git describe --tags --exact-match 2> /dev/null`
if [ $? -eq 0 ]; then
    echo $current_tag
else
    git branch | grep '^\*' | awk '{ print $2}'
fi
