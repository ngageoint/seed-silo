#!/usr/bin/env bash

sudo initctl stop silo
./build-silo.sh
sudo cp silo /bin/
sudo cp silo.conf /etc/init/
sudo mkdir -p /usr/silo
sudo chmod 755 /usr/silo
sudo initctl start silo