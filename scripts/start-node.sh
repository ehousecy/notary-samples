#!/usr/bin/env sh

print(){
  echo "$1"
}
#stop geth background
pkill geth

print "starting Notary service ..."
print "starting ethereum node..."
#start geth background
nohup geth --http --http.addr 127.0.0.1 --http.port 8545 --ws --ws.port 3334 --nodiscover --dev --dev.period 2 --datadir /tmp/ &
print "geth started successfully..."

until test -e /tmp/geth.ipc
do
    echo "waiting node initialize..."
    sleep 2
done
print "initializing notary ethereum address"
# config with 2 eth for gas fee
./apply_eth.sh 0x71BE5a9044F3E41c74b7c25AA19B528dd6B9f387 2
echo "geth successfully started!"



# start fabric nodes
echo "starting fabric networks"
cd fabric && ./network-start.sh
echo "started fabric network"
