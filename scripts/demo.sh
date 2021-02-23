#!/usr/bin/env sh
#set -xv

sep="#"
function append_cell(){
    #对表格追加单元格
    #append_cell col0 "col 1" ""
    #append_cell col3
    local i
    for i in "$@"
    do
        line+="|$i${sep}"
    done
}
function check_line(){
if [ -n "$line" ]
then
    c_c=$(echo $line|tr -cd "${sep}"|wc -c)
    difference=$((${column_count}-${c_c}))
    if [ $difference -gt 0 ]
    then
        line+=$(seq -s " " $difference|sed -r s/[0-9]\+/\|${sep}/g|sed -r  s/${sep}\ /${sep}/g)
    fi
    content+="${line}|\n"
fi

}
function append_line(){
    check_line
    line=""
    local i
    for i in "$@"
    do
        line+="|$i${sep}"
    done
    check_line
    line=""
}
function segmentation(){
    local seg=""
    local i
    for i in $(seq $column_count)
    do
        seg+="+${sep}"
    done
    seg+="${sep}+\n"
    echo $seg
}
function set_title(){
    #表格标头，以空格分割，包含空格的字符串用引号，如
    #set_title Column_0 "Column 1" "" Column3
#    [ -n "$title" ] && echo "Warring:表头已经定义过,重写表头和内容"
    column_count=0
    title=""
    local i
    for i in "$@"
    do
        title+="|${i}${sep}"
        let column_count++
    done
    title+="|\n"
    seg=`segmentation`
    title="${seg}${title}${seg}"
    content=""
}
function output_table(){
    if [ ! -n "${title}" ]
    then
        echo "未设置表头，退出" && return 1
    fi
    append_line
    table="${title}${content}$(segmentation)"
    echo -e $table|column -s "${sep}" -t|awk '{if($0 ~ /^+/){gsub(" ","-",$0);print $0}else{gsub("\\(\\*\\)","\033[31m(*)\033[0m",$0);print $0}}'

}


#generate accounts
fromInfo=$(../build/notary-cli gen-account)
# fromAddress=${fromInfoF:29:42}
substr=${fromInfo#*Address:}
address=${substr:0:45}
fromAddress=${address}
fromPriv=${fromInfo:85}

# generate receiver address
toInfo=$(../build/notary-cli gen-account)
toAddress=${toInfo:29:42}
toBalance=$(../build/notary-cli account --account $toAddress --network-type ethereum)

#get fabric accounts
fabricAccount=$(./fabric-account.sh)
fAAddress=$(echo $fabricAccount|jq -r ".org1")
fBAddress=$(echo $fabricAccount|jq -r ".org2")

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
Alice_MSP_HOME=$HOME/.notary-samples/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
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
