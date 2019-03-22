#!/bin/bash

PKG_FOLDER='pkg'

# Change to project directory and source common.sh
cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null 2>&1 && pwd)"
. 'scripts/common.sh'

for dir in $PKG_FOLDER/*; do
    pkg="${dir:$((${#PKG_FOLDER} + 1))}"
    echo_title "Testing $pkg"
    cd $PKG_FOLDER/$pkg
    go test
    if (($? > 0)); then
        echo_error "Some tests failed in $pkg"
    else
        echo_success "Tests of $pkg ran successfully"
    fi
    cd - > /dev/null
done
