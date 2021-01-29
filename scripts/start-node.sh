#!/usr/bin/env sh

print(){
  echo "$1"
}

print "starting Notary service ..."
print "starting ethereum node..."
#start geth background
nohup geth --http --http.addr 127.0.0.1 --http.port 8545 --ws --ws.port 3334 --nodiscover --dev &
print "geth started successfully..."

echo "geth successfully started!"



# start fabric nodes
echo "starting fabric networks"
cd fabric && ./network-start.sh
echo "started fabric network"
