#!/usr/bin/env sh

pkill geth
rm ~/.ethereum -rf


#stop fabric networks
cd ../fabric/fabric-samples/test-network && ./network.sh down
