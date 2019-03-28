#!/bin/bash

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

for dep in $(grep -rPo "(?<=\")github\.com.*(?=\"$)" . | cut -d':' -f2 | sort | uniq); do
    echo "Getting $dep"
    GOPATH="$GOPATH" go get -v -t "$dep"
done
