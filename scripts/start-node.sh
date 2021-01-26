#!/usr/bin/env sh

print(){
  echo "$1"
}

print "starting Notary service ..."
print "starting ethereum node..."
print "initing geth data"
geth init genesis.json
#start geth background
nohup spawn geth --http --http.addr 127.0.0.1 --http.port 8545 --ws --ws.port 3334 --nodiscover &
print "geth started successfully..."

#genenrate coinbase account
./gen-accounts.sh
./auto-mining.sh


