#!/bin/bash

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

GOPATH="$GOPATH" go run "cmd/$1/main.go"
