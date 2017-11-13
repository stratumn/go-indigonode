#!/bin/sh

# Starts Alice within a Docker container, first making sure configuration files
# exists.

set -e

if [ ! -f /usr/local/etc/alice/alice.core.toml ]; then
    echo Creating config file...

    # So that paths are within mountable dir.
    cd /usr/local/var/alice

    alice --core-config /usr/local/etc/alice/alice.core.toml --cli-config /usr/local/etc/alice/alice.cli.toml init
fi

echo Starting Alice...
exec alice --core-config /usr/local/etc/alice/alice.core.toml up
