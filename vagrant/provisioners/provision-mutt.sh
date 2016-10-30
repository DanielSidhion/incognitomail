#!/bin/bash
set -e

# Stuff needed to build mutt
sudo apt-get install -y build-essential libncurses5-dev libssl-dev

# libiconv has to be built to build mutt
wget -q -O libiconv-1.14.tar.gz http://ftp.gnu.org/pub/gnu/libiconv/libiconv-1.14.tar.gz
tar -xf libiconv-1.14.tar.gz
cd libiconv-1.14/srclib

# Apply libiconv patch to allow it to be built
patch < /vagrant/patches/fix_gets.patch

# Build libiconv
cd ..
./configure --prefix=/usr/local && make && sudo make install
sudo ldconfig

# Downloading mutt
cd /home/vagrant
wget -q -O mutt-1.7.1.tar.gz ftp://ftp.mutt.org/pub/mutt/mutt-1.7.1.tar.gz
tar -xf mutt-1.7.1.tar.gz
cd mutt-1.7.1

# Build mutt
./configure --enable-imap --with-ssl && make && sudo make install

# Adding mutt config to current user
sudo cp /vagrant/configs/mutt/muttrc /home/vagrant/.muttrc
sudo chown vagrant:vagrant /home/vagrant/.muttrc

# Remove all working files
cd /home/vagrant
rm -rf libiconv-1.14*
rm -rf mutt-1.7.1*
