#!/bin/bash
cd $(dirname $0) || exit 1
dir=$(pwd)
function fabricAccount() {
  # set org1 peer0 evn
  cd $dir/fabric/fabric-samples/test-network || exit 1
  export PATH=${PWD}/../bin:$PATH
  export FABRIC_CFG_PATH=$PWD/../config/

  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_ADDRESS=localhost:7051
  # query org1 admin account
  org1admin=$(peer chaincode query -C mychannel -n basic -c '{"Args":["ClientAccountID"]}')
  res=$?
  if [ $res -ne 0 ]; then
    echo "Failed to query fabric org1 admin account id..."
    exit 1
  fi
  echo "org1 admin account: $org1admin"

  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
  export CORE_PEER_ADDRESS=localhost:9051
  # query org2 admin account
  org2admin=$(peer chaincode query -C mychannel -n basic -c '{"Args":["ClientAccountID"]}')
  res=$?
  if [ $res -ne 0 ]; then
    echo "Failed to query fabric org2 admin account id..."
    exit 1
  fi
  echo "org2 admin account: $org2admin"
}

function ethAccount() {
  cd $dir/../build
  res=$?
  if [ $res -ne 0 ]; then
    echo "Make sure you execute the make build command successfully first..."
    exit 1
  fi

  if [ ! -f ./notary-cli ]; then
    echo "Make sure you execute the make build command successfully first..."
    exit 1
  fi
  count=$1
  if [ "$count" -lt 1 ]; then
    count=1
  fi
  int=1
  while (($int <= $count)); do
    ./notary-cli gen-account
    res=$?
    if [ $res -ne 0 ]; then
      echo "Failed to generate eth account..."
      exit 1
    fi
    (("int++"))
  done

}

function printHelp() {
  USAGE="$1"
  if [ "$USAGE" = "fabric" ]; then
    echo -e "Usage: "
    echo -e "  demo-account.sh \033[0;32mfabric\033[0m - Return org1 and org2 admin account id"
  elif [ "$USAGE" = "eth" ]; then
    echo -e "Usage: "
    echo -e "  demo-account.sh \033[0;32meth\033[0m [Flags]"
    echo -e ""
    echo -e "    Flags:"
    echo -e "    -n <account number> -  Number of eth accounts generated"
    echo -e " Examples:"
    echo -e "   demo-account.sh eth"
    echo -e "   demo-account.sh eth -n 2"
  else
    echo -e "Usage: "
    echo -e "  network.sh <Mode>"
    echo -e "    Modes:"
    echo -e "      \033[0;32mfabric\033[0m - Return org1 and org2 admin account id"
    echo -e "      \033[0;32meth\033[0m - Generate a new eth account(contain Address and Private Key)"
    echo -e ""
    echo -e "    Flags:"
    echo -e "    Used with \033[0;32mdemo-account.sh eth\033[0m"
    echo -e "    -n <account number> -  Number of eth accounts generated"
    echo -e ""
    echo -e " Examples:"
    echo -e "   demo-account.sh fabric"
    echo -e "   demo-account.sh eth"
    echo -e "   demo-account.sh eth -n 2"
  fi
}

ACCOUNT_NUM=1

## Parse mode
if [[ $# -lt 1 ]]; then
  printHelp
  exit 0
else
  MODE=$1
  shift
fi

while [[ $# -ge 1 ]]; do
  key="$1"
  if [ "$key" == "-h" ]; then
    printHelp $MODE
    exit 0
  elif [ "$key" == "-n" ]; then
    # 判断是否为数值
    echo "$2" | grep "[^0-9]" >/dev/null
    res=$?
    if [ $res -ne 1 ]; then
      echo -e "\033[0;31mError -n flag value: $2\033[0m"
      printHelp
      exit 1
    fi
    ACCOUNT_NUM="$2"
    shift
  else
    echo -e "\033[0;31mUnknown flag: $key\033[0m"
    printHelp
    exit 1
  fi
  shift
done

if [ "$MODE" == "fabric" ]; then
  fabricAccount
elif [ "$MODE" == "eth" ]; then
  ethAccount $ACCOUNT_NUM
else
  printHelp
fi
