#!/usr/bin/env sh

pkill geth
rm -rf $HOME/monitor
#stop fabric networks
cd fabric && ./network-stop.sh

#rm notary-server data
if [ -f $HOME/.notary-samples/foo.db ]; then
    rm -f $HOME/.notary-samples/foo.db
fi
