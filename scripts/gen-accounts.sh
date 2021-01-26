#!/usr/bin/expect

set timeout 9999
spawn geth account new
expect "*Password: " { send "123\r" }
expect "Repeat password:"  { send "123\r" }
expect eof