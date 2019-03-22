#!/bin/bash

CMD_FOLDER='cmd'
BUILD_FOLDER='build'

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null 2>&1 && pwd)"
. 'scripts/common.sh'

for dir in $CMD_FOLDER/*; do
    pkg="${dir:$((${#CMD_FOLDER} + 1))}"
    echo_title "Building $pkg"
    builddir="$BUILD_FOLDER/$pkg"
    if [[ -d "$builddir" ]]; then
        echo "Deleting existing build folder $builddir"
        rm -r "$builddir"
    fi
    mkdir -p "$BUILD_FOLDER"
    mainfile="$CMD_FOLDER/$pkg/main.go"
    go build -o "$builddir" "$mainfile"
    if (($? > 0)); then
        echo_error "Failed building $pkg"
    fi
done
