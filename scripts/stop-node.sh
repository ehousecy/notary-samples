#!/usr/bin/env sh

pkill geth
rm ~/.ethereum -rf


#stop fabric networks
cd fabric && ./network-stop.sh
