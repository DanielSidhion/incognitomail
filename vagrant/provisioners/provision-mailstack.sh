#!/bin/bash

sudo cp /vagrant/configs/general/hosts /etc/hosts
sudo chown root:root /etc/hosts

sudo apt-get update

# Selecting postfix configurations to allow automatic installation
sudo echo "postfix postfix/main_mailer_type select Internet Site" > debconf-selections.txt
sudo echo "postfix postfix/mailname string mail.incognitotest" >> debconf-selections.txt
sudo debconf-set-selections debconf-selections.txt
sudo apt-get install -y mail-stack-delivery
sudo rm debconf-selections.txt

sudo apt-get install -y mailutils

# Copying postfix configuration files and updating postfix's databases
sudo cp /vagrant/configs/postfix/canonical /etc/postfix/canonical
sudo cp /vagrant/configs/postfix/main.cf /etc/postfix/main.cf
sudo cp /vagrant/configs/postfix/master.cf /etc/postfix/master.cf
sudo cp /vagrant/configs/postfix/virtual_aliases /etc/postfix/virtual_aliases
sudo cp /vagrant/configs/postfix/virtual_mailbox_domains /etc/postfix/virtual_mailbox_domains
sudo cp /vagrant/configs/postfix/virtual_mailbox_users /etc/postfix/virtual_mailbox_users
sudo chown -R root:root /etc/postfix

sudo newaliases
sudo postmap /etc/postfix/virtual_aliases
sudo postmap /etc/postfix/canonical
sudo postmap /etc/postfix/virtual_mailbox_domains
sudo postmap /etc/postfix/virtual_mailbox_users

# Copying dovecot's configuration files
sudo cp /vagrant/configs/dovecot/99-mail-stack-delivery.conf /etc/dovecot/conf.d/99-mail-stack-delivery.conf
sudo cp /vagrant/configs/dovecot/passwd.db /etc/dovecot/passwd.db
sudo chown -R postfix:postfix /etc/dovecot

# User used for creating vmail boxes
sudo groupadd -g 5000 vmail
sudo useradd -g vmail -u 5000 vmail -d /var/mail/vmail -m

# Restarting both services to update changes
sudo systemctl restart dovecot
sudo systemctl restart postfix
