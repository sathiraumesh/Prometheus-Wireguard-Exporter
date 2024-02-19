#!/bin/bash

apt-get update 

apt-get install -y wireguard 

apt-get install -y iproute2

apt-get install -y nmap

apt-get install -y curl

cd /etc/wireguard && mkdir keys && cd keys

# generate random keys for conf
# wg genkey | tee privatekey | wg pubkey > publickey