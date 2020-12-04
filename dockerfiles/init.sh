#!/bin/bash

apt-get update 
apt-get install -y nasm clang llvm

curl -L https://golang.org/dl/go1.15.6.linux-amd64.tar.gz | tar -C /usr/local -xzf -
ln -s /usr/local/go/bin/go /usr/local/bin/go
