#!/usr/bin/env sh
echo "[process] stop Fabric testnet"

cd fabric-samples/test-network
./network.sh down

echo "[finished] stop Fabric testnet"
