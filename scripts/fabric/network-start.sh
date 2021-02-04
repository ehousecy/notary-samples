#!/bin/bash
echo "====================== [process] start Fabric docker network ========================="

if [ ! -d fabric-samples ]; then
  curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.1 1.4.9
fi

cd fabric-samples/test-network || exit
./network.sh down
./network.sh up
res=$?
if [ $res -ne 0 ]; then
    fatalln "Failed to up fabric network..."
fi

sleep 1
./network.sh createChannel
res=$?
if [ $res -ne 0 ]; then
    fatalln "Failed to create fabric channel..."
fi
sleep 1
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
res=$?
if [ $res -ne 0 ]; then
    fatalln "Failed to deploy fabric chaincode..."
fi
sleep 1

cd ../..
./network-init.sh

#cp msp to server fabric business
if [ ! -d "$HOME"/.notary-samples ]; then
    if [ -f "$HOME"/.notary-samples ]; then
        rm -f "$HOME"/.notary-samples
        res=$?
        if [ $res -ne 0 ]; then
            fatalln "Failed to rm $HOME/.notary-samples file..."
        fi
    fi
    mkdir "$HOME"/.notary-samples
    res=$?
        if [ $res -ne 0 ]; then
            fatalln "Failed to exec mkdir $HOME/.notary-samples dir..."
        fi
fi
cp -r fabric-samples/test-network/organizations "$HOME"/.notary-samples/
res=$?
if [ $res -ne 0 ]; then
    fatalln "Failed to cp msp for server fabric business..."
fi
echo "====================== [finished] start Fabric docker network ========================="
