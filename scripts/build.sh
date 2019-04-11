#!/bin/bash

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

CMD_FOLDER="$ROOT/cmd"
BUILD_FOLDER="$GOPATH/bin"

for dir in $CMD_FOLDER/*; do
    pkg="${dir:$((${#CMD_FOLDER} + 1))}"
    echo_title "Building $pkg"
    GOPATH="$GOPATH" go build -o "$BUILD_FOLDER/$pkg" -i "$CMD_FOLDER/$pkg/main.go"
    if (($? > 0)); then
        echo_error "Failed building $pkg"
    fi
done
