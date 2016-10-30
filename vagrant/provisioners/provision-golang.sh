#!/bin/bash

GOLANG_VERSION=1.7.3
GOLANG_PACKAGE=go${GOLANG_VERSION}.linux-amd64.tar.gz

# Fetching golang binaries
wget -q -O /usr/local/src/${GOLANG_PACKAGE} https://storage.googleapis.com/golang/${GOLANG_PACKAGE}
tar -C /usr/local -xf /usr/local/src/${GOLANG_PACKAGE}

# Checking if go binaries are already in path
grep -q /usr/local/go /home/vagrant/.profile

if test $? -eq 1; then
	echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/vagrant/.profile
fi

# Checking if go workspace is already in path
grep -q GOPATH /home/vagrant/.profile

if test $? -eq 1; then
	echo 'export GOPATH=/home/vagrant/go' >> /home/vagrant/.profile
fi
