#!/usr/bin/env sh

# download and install geth
GOPROXY="https://goproxy.io,direct"
echo "download and installing ethereum binary"
go get -v github.com/ethereum/go-ethereum/cmd/geth

if [ $? -ne 0 ];then
  echo "download ethereum binary failed"
  exit -1
else
  echo "successfully downloaded ethereum binary"
fi

# download and install fabric binary

echo  "installing fabric binaries"
mkdir -p ./fabric
cd fabric
#curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.1 1.4.9
# replace shot url with raw url
curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash -s -- 2.2.1 1.4.9

if [ $? -ne 0 ];then
  echo "download fabric binary failed"
  exit -1
else
  echo "successfully downloaded fabric binary"
fi