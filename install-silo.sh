#!/usr/bin/env bash

build-silo.sh
sudo cp silo /bin/
sudo cp silo.conf /etc/init/
sudo mkdir /usr/silo
sudo chmod 755 /usr/silo
sudo initctl stop silo
sudo initctl start silo