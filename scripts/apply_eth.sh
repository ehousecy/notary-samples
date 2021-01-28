#!/usr/bin/expect
set timeout 9999
set to [lindex $argv 0];
set amount [lindex $argv 1]
spawn geth attach  /tmp/geth.ipc
expect "* press ctrl-d*" { send "eth.sendTransaction({from:eth.coinbase, to:\"$to\", amount: $amount})\r"}
expect "*0x*" { send "exit\r" }
expect eof