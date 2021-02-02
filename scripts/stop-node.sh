#!/usr/bin/env sh

pkill geth

#stop fabric networks
cd fabric && ./network-stop.sh
