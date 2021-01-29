#!/bin/bash
echo "[process] stop Fabric network"

cd fabric-samples/test-network || exit
./network.sh down

echo "[finished] stop Fabric network"
