#!/bin/bash

source /home/vagrant/.profile

go get github.com/danielsidhion/incognitomail/...
sudo chown -R vagrant:vagrant /home/vagrant/go
