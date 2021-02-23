#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# This script brings up a Hyperledger Fabric network for testing smart contracts
# and applications. The test network consists of two organizations with one
# peer each, and a single node Raft ordering service. Users can also use this
# script to create a channel deploy a chaincode on the channel
#
# prepending $PWD/../bin to PATH to ensure we are picking up the correct binaries
# this may be commented out to resolve installed version of tools if desired
echo "====================== [process] init Fabric network data ========================="
cd fabric-samples/test-network || exit
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"Mint","Args":["3000"]}'
res=$?
if [ $res -ne 0 ]; then
  echo "Failed to init fabric chaincode data..."
  exit 1
fi
sleep 3

org1admin=$(peer chaincode query -C mychannel -n basic -c '{"Args":["ClientAccountID"]}')
res=$?
if [ $res -ne 0 ]; then
  echo "Failed to query fabric chaincode data..."
  exit 1
fi
echo "org1admin account id: $org1admin"

#get org2 admin account id
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051
org2admin=$(peer chaincode query -C mychannel -n basic -c '{"Args":["ClientAccountID"]}')
res=$?
if [ $res -ne 0 ]; then
  echo "Failed to query fabric chaincode data..."
  exit 1
fi
echo "org2admin account id: $org2admin"

#org1 transfer balance to org2
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"Transfer","Args":["'$org2admin'","1000"]}'
res=$?
if [ $res -ne 0 ]; then
  echo "Failed to init fabric chaincode data..."
  exit 1
fi
sleep 3
org1adminBalance=$(peer chaincode query -C mychannel -n basic -c '{"Args":["BalanceOf","'$org1admin'"]}')
res=$?
if [ $res -ne 0 ]; then
  echo "Failed to query fabric chaincode data..."
  exit 1
fi
echo "org1admin account Balance: $org1adminBalance"
org2adminBalance=$(peer chaincode query -C mychannel -n basic -c '{"Args":["BalanceOf","'$org2admin'"]}')
res=$?
if [ $res -ne 0 ]; then
  echo "Failed to query fabric chaincode data..."
  exit 1
fi
echo "org2admin account Balance: $org2adminBalance"

echo "====================== [finished] init Fabric network data ========================="
