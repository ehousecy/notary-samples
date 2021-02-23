#!/bin/bash
cd $(dirname $0) || exit 1
dir=$(pwd)
cd $dir/fabric/fabric-samples/test-network || exit 1
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
org1admin=$(peer chaincode query -C mychannel -n basic -c '{"Args":["ClientAccountID"]}')
res=$?
if [ $res -ne 0 ]; then
    echo "Failed to query fabric org1 admin account id..."
    exit 1
fi
#echo "org1 admin account: $org1admin"

export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051
org2admin=$(peer chaincode query -C mychannel -n basic -c '{"Args":["ClientAccountID"]}')
res=$?
if [ $res -ne 0 ]; then
    echo "Failed to query fabric org2 admin account id..."
    exit 1
fi
#echo "org2 admin account: $org2admin"
echo "{\"org1\":\"$org1admin\",\"org2\":\"$org2admin\"}"
