#!/bin/bash
echo "====================== [process] start docker environment for Fabric testnet ========================="

if [ ! -d fabric-samples ]; then
  curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.1 1.4.9
fi

cd fabric-samples/test-network || exit
./network.sh down
./network.sh up

sleep 1
./network.sh createChannel
sleep 1
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
sleep 1

cd ../..
./network-init.sh

#cp msp to server fabric business
if [ ! -d "$HOME"/.notary-samples ]; then
    mkdir "$HOME"/.notary-samples
fi
cp -r fabric-samples/test-network/organizations "$HOME"/.notary-samples/
echo "====================== [finished] start docker environment for Fabric testnet ========================="
