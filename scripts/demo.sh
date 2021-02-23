#!/usr/bin/env bash
#set -xv

. utils.sh

printf "\n==============================initializing==============================\n\n"
#generate accounts
fromInfo=$(./demo-account.sh eth)
# fromAddress=${fromInfoF:29:42}
substr=${fromInfo#*Address:}
address=${substr:2:42}
fromAddress=${address}
fromPriv=${fromInfo:85}

# generate receiver address
toInfo=$(./demo-account.sh eth)
toAddress=${toInfo:29:42}

#get fabric accounts
fabricAccount=$(./demo-account.sh fabric)
fAAddress=${fabricAccount#*account: }
fAAddress=${fAAddress%org2*}
fBAddress=${fabricAccount##*account: }

# apply ethereum from faucet and display sender address info

printf "\n==============================initializing==============================\n\n"
printf "applying eth for Alice...\n"
./apply_eth.sh $fromAddress 100
sleep 5
printf "applied 100eth for alice\n"

# query blockchain info
fromBalance=$(../build/notary-cli account --account $fromAddress --network-type ethereum)
toBalance=$(../build/notary-cli account --account $toAddress --network-type ethereum)
fABalance=$(../build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $fAAddress)
fBBalance=$(../build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $fBAddress)
# dispaly user account info
printf -- "\n-----------------------------Alice Account info-----------------------------\n"
#printf "%8s %43s %s\n" "network" "account" "balance"
#printf "%8s %43s %.2f\n" "ethereum" "$fromAddress" "$fromBalance"
#printf "%8s %43s %.2f\n" "fabric" "$fAAddress" "$fABalance"
set_title Network Account Balance
append_line ethereum "$fromAddress" "$fromBalance"
append_cell fabric "$fAAddress" "$fABalance"
output_table

printf -- "\n-----------------------------Bob Account info-----------------------------\n"
#printf "%8s %43s %s\n" "network" "account" "balance"
#printf "%8s %43s %.2f\n" "ethereum" "$toAddress" "$toBalance"
#printf "%8s %43s %.2f\n" "fabric" "$fBAddress" "$fBBalance"
set_title Network Account Balance
append_line ethereum "$toAddress" "$toBalance"
append_cell fabric "$fBAddress" "$fBBalance"
output_table


#start cross chain process
printf "\n==============================transferring asset across blockchain==============================\n\n"
ticket=$(../build/notary-cli create-ticket --efrom $fromAddress --eto $toAddress --eamount 10 --ffrom $fBAddress --fto $fAAddress --famount 100 --fchannel mychannel --fcc basic 2>&1)
id=${ticket#*ticketId:}
id=$(echo $id |grep -o  "[0-9]*")
if [ "$id" == "" ];then
  printf "\nfailed to create cross-chain ticket\n"
  exit 1
fi
printf "\n\t 1. created cross-chain ticket, ID: %s\n" "$id"
printf "\t    [info]:Alice tranfering 10 eth for Bob 100 fabric assets\n"

#submit transactions

printf "\t 2. submitting fabric transaction\n"
printf "\t   [info]: sending 100 fabric asset from Bob to Notary address\n"
Bob_MSP_HOME=$HOME/.notary-samples/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
cmdResp=$(../build/notary-cli submit --network-type fabric --msp-path $Bob_MSP_HOME --msp-id Org2MSP --ticket-id $id 2>&1)
if [ $? -ne 0 ];then
  printf "failed to submit fabric transaction\n"
  exit 1
fi

printf "\t 3. submitting ethereum transaction\n"
printf "\t   [info]: sending 10 eth from Alice to Notary address\n"
resp=$(../build/notary-cli submit --network-type ethereum --private-key $fromPriv --ticket-id $id 2>&1)

if [ $? -ne 0 ];then
  printf "failed to submit ethereum transaction\n"
  exit 1
fi


#wait for 6 block confirmation
printf "\t   [info]: waiting for ethereum network confirm tx\n"
sleep 16

#approve cross-chain ticket
printf "\t 4. approving cross chain ticket\n"
printf "\t   [info]:transfering ethereum and fabric assets from Notary to the finally receiver\n"
cmdResp=$(../build/notary-cli approve --ticket-id $id 2>&1)
printf "\t   [info]:successfully approved cross-chain ticket\n"
#display blockchain properties
# wait for a new block
printf "\t   [info]: waiting for blockchain confirm transactions...\n"
sleep 6
printf "\n==============================finished transfer assets==============================\n\n"
# query blockchain info
fromBalance=$(../build/notary-cli account --account $fromAddress --network-type ethereum)
toBalance=$(../build/notary-cli account --account $toAddress --network-type ethereum)
fABalance=$(../build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $fAAddress)
fBBalance=$(../build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $fBAddress)
# dispaly user account info
printf -- "\n-----------------------------Alice Account info-----------------------------\n"
#printf "%8s %43s %s\n" "network" "account" "balance"
#printf "%8s %43s %.2f\n" "ethereum" "$fromAddress" "$fromBalance"
#printf "%8s %43s %.2f\n" "fabric" "$fAAddress" "$fABalance"
set_title Network Account Balance
append_line ethereum "$fromAddress" "$fromBalance"
append_cell fabric "$fAAddress" "$fABalance"
output_table

printf -- "\n-----------------------------Bob Account info-----------------------------\n"
#printf "%8s %43s %s\n" "network" "account" "balance"
#printf "%8s %43s %.2f\n" "ethereum" "$toAddress" "$toBalance"
#printf "%8s %43s %.2f\n" "fabric" "$fBAddress" "$fBBalance"
set_title Network Account Balance
append_line ethereum "$toAddress" "$toBalance"
append_cell fabric "$fBAddress" "$fBBalance"
output_table
