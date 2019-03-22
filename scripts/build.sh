#!/bin/bash

CMD_FOLDER='cmd'
BUILD_FOLDER='build'

# Get the git commit
GIT_COMMIT=$(git rev-parse HEAD 2> /dev/null)
if (( $? > 0 )); then
    echo "Not inside git repository." >&2
    exit 1
fi
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
echo "Commit: $GIT_COMMIT$GIT_DIRTY"

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo "Go not installed" >&2
    exit 1
fi

# If inside script folder, change to project folder
if [[ "${PWD##*/}" == "scripts" ]]; then
    echo "Inside scripts folder. Changing to project folder."
    cd ..
fi

# Build each main.go within the folders in $CMD_FOLDER into $BUILD_FOLDER
for dir in $CMD_FOLDER/*; do
    pkg="${dir:$((${#CMD_FOLDER} + 1))}"
    echo "Building $pkg"
    builddir="$BUILD_FOLDER/$pkg"
    if [[ -d "$builddir" ]]; then
        echo "Deleting existing build folder $builddir"
        rm -r "$builddir"
    fi
    mkdir -p "$BUILD_FOLDER"
    mainfile="$CMD_FOLDER/$pkg/main.go"
    go build -o "$builddir" "$mainfile"
done
