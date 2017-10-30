#!/usr/bin/env bash

sudo initctl stop silo
build-silo.sh
sudo cp silo /bin/
sudo cp silo.conf /etc/init/
sudo initctl start silo