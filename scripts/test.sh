#!/bin/bash

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

PKG_FOLDER='pkg'

for dir in $PKG_FOLDER/*; do
    pkg="${dir:$((${#PKG_FOLDER} + 1))}"
    echo_title "Testing $pkg"
    GOPATH="$GOPATH" go test $PKG_FOLDER/$pkg/*
    if (($? > 0)); then
        echo_error "Some tests failed in $pkg"
    else
        echo_success "Tests of $pkg ran successfully"
    fi
done
