## 下载项目

```bash
git clone https://github.com/ehousecy/notary-samples.git
```

## 测试区块链网络环境搭建

```bash
#进入项目
cd notary-samples
#安装以太坊和fabric网络文件
make install
#启动以太坊和fabric网络，启动前确保安装expect，命令：apt-get install expect
make start
```

## 准备跨链材料

```bash
#编译跨链server端和cli端
make build

#准备以太坊账户
##转账账户
./build/notary-cli gen-account
#result
#Address:  0xB696AaF5ea7455a65Be5a765c9b9F2e351B60a09
#Private Key:  c25a485cefa7ff54b29680c477fc89c4ccb16ca975fd861af864ae6b63227000

##接收账户
./build/notary-cli gen-account
#result
#Address:  0x2B7B8Dc4c646613AA55BB13b0Ec09232692677D4
#Private Key:  324d34b4c06c52cf995cf759878f6c93615eedea30d491260c7651426586f900

#转账账户申请以太
./scripts/apply_eth.sh 0xB696AaF5ea7455a65Be5a765c9b9F2e351B60a09 1000

#重新打开一个terminal,进入notary-samples目录,执行命令启动跨链服务
./build/notary-server
```

## 查询跨链前资产信息

```bash
# 以太账户资产信息
./build/notary-cli account --network-type ethereum --account 0xB696AaF5ea7455a65Be5a765c9b9F2e351B60a09

./build/notary-cli account --network-type ethereum --account 0x2B7B8Dc4c646613AA55BB13b0Ec09232692677D4

# fabric资产信息
./build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account asset1
```

## 执行跨链流程

```bash
#1.创建跨链交易
./build/notary-cli create-ticket --efrom 0xB696AaF5ea7455a65Be5a765c9b9F2e351B60a09 --eto 0x2B7B8Dc4c646613AA55BB13b0Ec09232692677D4 --eamount 10 --ffrom Tomoko --fto Max --famount asset1 --fchannel mychannel --fcc basic
#返回ticket-id信息

#2.用户提交fabric交易,--ticket-id参数填入步骤1返回的ticket-id值
MSP_HOME=$HOME/.notary-samples/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp && ./build/notary-cli submit --network-type fabric --msp-path $MSP_HOME --msp-id Org1MSP --ticket-id 1

#3.用户提交以太交易,--ticket-id参数填入步骤1返回的ticket-id值
./build/notary-cli submit --network-type ethereum --private-key c25a485cefa7ff54b29680c477fc89c4ccb16ca975fd861af864ae6b63227000 --ticket-id 1

#4.完成公证人转账交易
./build/notary-cli approve --ticket-id 1

#5.查询跨链交易信息
./build/notary-cli query --ticket-id 1
```

## 查询跨链后资产信息

```bash
# 以太账户资产信息
./build/notary-cli account --network-type ethereum --account 0xB696AaF5ea7455a65Be5a765c9b9F2e351B60a09

./build/notary-cli account --network-type ethereum --account 0x2B7B8Dc4c646613AA55BB13b0Ec09232692677D4

# fabric资产信息
./build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account asset1
```

