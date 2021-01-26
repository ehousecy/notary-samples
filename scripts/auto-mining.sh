#!/usr/bin/expect

set timeout 9999
spawn geth attach $::env(HOME)/.ethereum/geth.ipc
expect "*press ctrl-d*" { send "miner.start()\r"}
expect "null" { send "exit\r" }
expect eof