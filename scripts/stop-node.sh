#!/usr/bin/env sh

pkill geth

#stop fabric networks
cd fabric && ./network-stop.sh

#rm notary-server data
if [ -f $HOME/.notary-samples/foo.db ]; then
    rm -f $HOME/.notary-samples/foo.db
fi
