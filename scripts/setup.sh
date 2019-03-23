#!/bin/bash

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

GOPATH="$GOPATH" go get -d ./...
