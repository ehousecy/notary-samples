#!/usr/bin/env sh

print(){
  echo "$1"
}

print "starting Notary service ..."
print "starting ethereum node..."
print "initing geth data"
geth init genesis.json
#start geth background
echo "generating coinbase"
./gen-accounts.sh
nohup geth --http --http.addr 127.0.0.1 --http.port 8545 --ws --ws.port 3334 --nodiscover &
print "geth started successfully..."

until test -e ~/.ethereum/geth.ipc
do
    echo "waiting node initialize..."
    sleep 2
done

#genenrate coinbase account
#echo "generating coinbase"
#./gen-accounts.sh
./auto-mining.sh

echo "geth successfully started!"



# start fabric nodes
echo "starting fabric networks"
cd fabric && ./network-start.sh
echo "started fabric network"
