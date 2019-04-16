#!/bin/bash

# Go to project root folder and set some variables
cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null 2>&1 && pwd)"
#echo $OSTYPE
#pwd
#case `pwd` in /home/*)
#    echo "Gotta change GOPATH variable"
#esac
#eval "$($(which go) env)"
#if [[ -z "$GOPATH" ]]; then
    GOPATH="$(cd "./../../../.." > /dev/null 2>&1 && pwd)"
    echo "Assuming GOPATH=$GOPATH"
#fi
ROOT=`pwd`

# Formatting
[[ -n "$GOLAND" ]] && echo "GOLAND variable set"
FG_BLACK="$(tput setaf 0)" #<- messes with the goland terminal
[[ -n "$GOLAND" ]] && FG_BLACK="\033[97m"
FG_RED="$(tput setaf 1)"
FG_GREEN="$(tput setaf 6)"
BG_BLUE="$(tput setab 4)"
RESET="$(tput sgr0)" #<- messes with the goland terminal
[[ -n "$GOLAND" ]] && RESET="\033[m"
function echo_title() { echo -e "$BG_BLUE$FG_BLACK $1 $RESET"; }
function echo_error() { echo -e "$FG_RED$1$RESET"; }
function echo_success() { echo -e "$FG_GREEN$1$RESET"; }

# Get the git commit
GIT_COMMIT=$(git rev-parse HEAD 2> /dev/null)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
echo_title "Commit: $GIT_COMMIT$GIT_DIRTY"

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo_error "Go not installed" >&2
    exit 1
fi
