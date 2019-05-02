#!/bin/bash

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

for dep in $(grep -rPo "(?<=\")github\.com.*(?=\"$)" . | cut -d':' -f2 | sort | uniq); do
    echo "Getting $dep"
    GOPATH="$GOPATH" go get -v -t "$dep"
done

echo "Creating test user(test:secret)"
HOMEDIR='/tmp/test'
sudo useradd test -d "$HOMEDIR" -p '$1$Qd8H95T5$RYSZQeoFbEB.gS19zS99A0' -s /bin/bash
sudo mkdir -p "$HOMEDIR"
sudo chown test "$HOMEDIR"
sudo chmod 755 "$HOMEDIR"

#https://blog.fabioiotti.com/generate-rsa-key/
# openssl genpkey -out mykey.pem -algorithm rsa -pkeyopt rsa_keygen_bits:2048
# openssl rsa -in mykey.pem -out mykey.pub -pubout