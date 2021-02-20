#!/usr/bin/env sh
#set -xv

demoPrint (){
  echo "[Info]" $1
}

#generate accounts
fromInfo=$(../build/notary-cli gen-account)
fromAddress=${fromInfoF:29:42}
fromPriv=${fromInfo:85}
toInfo=$(../build/notary-cli gen-account)
toAddress=${toInfo:29:42}
demoPrint "generating ethereum Accounts:"
demoPrint "generated sender account : $fromAddress"
demoPrint "sender account private key: $fromPriv"
demoPrint "generated receiver account: $toAddress"
toBalance=$(../build/notary-cli account --account $toAddress --network-type ethereum)
demoPrint "ethereum receiver address balance: $toBalance"

# apply ethereum from faucet and display sender address info
demoPrint "applying eth from faucet for sender account"
./apply_eth.sh $fromAddress 100
sleep 5
fromBalance=$(../build/notary-cli account --account $fromAddress --network-type ethereum)
demoPrint "ethereum sender account balance: $fromBalance"

# display fabric assets
fabBalance=$(../build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account asset1)
demoPrint "fabric assets: $fabBalance"

#start cross chain process
demoPrint "creating cross-chain ticket"
ticket=$(../build/notary-cli create-ticket --efrom $fromAddress --eto $toAddress --eamount 10 --ffrom Tomoko --fto Max --famount asset1 --fchannel mychannel --fcc basic 2>&1)
demoPrint "${ticket}"
#id=${info#*ticketId:}
id=${ticket#*ticketId:}
id=$(echo $id |grep -o  "[0-9]*")
if [ "$id" == "" ];then
  demoPrint "failed to create cross-chain ticket"
  exit 1
fi
demoPrint $ticket

#submit transactions
demoPrint "submitting fabric transaction"
MSP_HOME=$HOME/.notary-samples/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
../build/notary-cli submit --network-type fabric --msp-path $MSP_HOME --msp-id Org1MSP --ticket-id $id

if [ $? -ne 0 ];then
  demoPrint "failed to submit fabric transaction"
  exit 1
fi

demoPrint "submitting ethereum transaction"
../build/notary-cli submit --network-type ethereum --private-key $fromPriv --ticket-id $id

if [ $? -ne 0 ];then
  demoPrint "failed to submit ethereum transaction"
  exit 1
fi

#wait for 6 block confirmation
sleep 16

#approve cross-chain ticket
../build/notary-cli approve --ticket-id $id

#display blockchain properties
# wait for a new block
sleep 6

toBalance=$(../build/notary-cli account --network-type ethereum --account $toAddress)
demoPrint "$toAddress: $toBalance"

ftoBalance=$(../build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account asset1)

demoPrint "$ftoBalance"