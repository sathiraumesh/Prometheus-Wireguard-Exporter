#!/bin/bash

apt-get update

apt-get install -y wireguard iproute2

cd /etc/wireguard && mkdir keys && cd keys

# generate random keys for conf
wg genkey | tee privatekey | wg pubkey > publickey

exit 0
