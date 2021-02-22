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

### 编译跨链server端和cli端可执行文件

```shell
make build
```

如果执行成功会在当前目录下新建`build`目录，并存放`notary-server`和`notary-cli`可执行文件。

### 准备以太坊

使用以下命令来生成以太坊账户：

```shell
./build/notary-cli gen-account
```

执行成功将看到类似如下输出：

```shell
Generated Account:
Address:  0x8635A5A979F56CfBE310C79241F941A16D3d70c5
Private Key:  25e5f1a85a18292c7abbf629857e3cd3e01a762c38e790c332829ce912088400
```

输出中`Address`为以太坊地址，`Private Key`是对应的私钥，用于对交易签名。后续执行跨链交易时需要这两个信息，因此我们将其保存在环境变量中。

```shell
export AliceEthAcc=0x8635A5A979F56CfBE310C79241F941A16D3d70c5 && export AliceEthPK=25e5f1a85a18292c7abbf629857e3cd3e01a762c38e790c332829ce912088400
```

> Note:每次生成的以太账户是不同的，上述命令因填入`./build/notary-cli gen-account`的返回值

重复以上步骤生成一个接收资产以太坊账户

```shell
./build/notary-cli gen-account
```

输出：

```shell
Generated Account:
Address:  0xB696AaF5ea7455a65Be5a765c9b9F2e351B60a09
Private Key:  c25a485cefa7ff54b29680c477fc89c4ccb16ca975fd861af864ae6b63227000
```

设置接收资产以太坊账户环境变量

```shell
export BobEthAcc=0xB696AaF5ea7455a65Be5a765c9b9F2e351B60a09 && export BobEthPK=c25a485cefa7ff54b29680c477fc89c4ccb16ca975fd861af864ae6b63227000
```

执行以下命令为转账账户申请以太

```shell
./scripts/apply_eth.sh $AliceEthAcc 1000
```

### 准备fabric账户

执行以下命令查看fabric账户

```shell
./scripts/fabric-account.sh
```

执行成功将看到类似如下输出：

```shell
org1 admin account: 4FA09DFE101BE0816DBDAD53B48EA8A9
org2 admin account: D9342B4D91186221EC18C8723E2786E9
```

输出的账户信息通过证书转换获得，`org1 admin account`和`org2 admin account`分别对应Org1和Org2的admin证书。接着通过以下命令为fabric账户以及msp证书目录设置环境变量：

```shell
export AliceFabricAcc=4FA09DFE101BE0816DBDAD53B48EA8A9 && export Alice_MSP_HOME=$HOME/.notary-samples/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp

export BobFabricAcc=D9342B4D91186221EC18C8723E2786E9 && export Bob_MSP_HOME=$HOME/.notary-samples/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
```

## 启动跨链server服务

为了查看server服务执行的详细信息，重新打开一个terminal，执行启动命令

```shell
#进入项目目录
cd notary-samples
#启动跨链服务
./build/notary-server
```

执行成功将在控制台中打印fabric和以太坊网络中区块监听信息。

如果不想切换terminal，可以使用以下命令后台启动跨链服务:

```shell
nohup ./build/notary-server &
```

## 查询跨链前资产信息

```bash
# 以太账户资产信息
./build/notary-cli account --network-type ethereum --account $AliceEthAcc

./build/notary-cli account --network-type ethereum --account $BobEthAcc

# fabric资产信息
./build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $AliceFabricAcc
./build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $BobFabricAcc
```

## 执行跨链流程

### 1.创建跨链交易

```shell
./build/notary-cli create-ticket --efrom $AliceEthAcc --eto $BobEthAcc --eamount 10 --ffrom $AliceFabricAcc --fto $BobFabricAcc --famount 100 --fchannel mychannel --fcc basic
```

如果创建跨链交易成功，将会看到类似以下输出：

```shell
2021/02/21 12:24:32 Successfully created ticket!
2021/02/21 12:24:32 ticketId:"1"
```

其中`ticketId`为跨链交易的唯一标识，后续流程都会使用到此标识，因此为设置环境变量`ticketId`：

```shell
export ticketId=1
```

### 2.用户提交以太交易

```shell
./build/notary-cli submit --network-type ethereum --private-key $AliceEthPK --ticket-id $ticketId
```

交易提交后跨链server服务会打印类似以下输入确认交易落块：

```shell
2021/02/21 15:26:00 [Eth handler] received ethereum transaction, id: 0x2f427d1be58abc190fc9fd5a79a87d4efac7a13e967202b2cf6c458c37f4b1a3
...
2021/02/21 15:26:16 [Eth handler] send out tx confirm event, tx id: 0x2f427d1be58abc190fc9fd5a79a87d4efac7a13e967202b2cf6c458c37f4b1a3
2021/02/21 15:26:16 [Eth handler] Confirmed tx, removing tx: 0x2f427d1be58abc190fc9fd5a79a87d4efac7a13e967202b2cf6c458c37f4b1a3
```

### 3.用户提交fabric交易

```shell
./build/notary-cli submit --network-type fabric --msp-path $Alice_MSP_HOME --msp-id Org1MSP --ticket-id $ticketId
```

交易提交后跨链server服务会打印类似以下输入确认交易落块：

```shell
2021/02/21 15:26:25 开始 fabric 交易, ticketID:1
2021/02/21 15:26:25 处理 fabric 交易：proposal签名完成, ticketID:1
2021/02/21 15:26:25 处理 fabric 交易：交易签名完成, ticketID:1, txID:d0a2ee339923cc9739bc7900765cb29b72f0a73c7876306fadd7d5ba79a0f332
2021/02/21 15:26:25 发送 fabric 交易, ticketID:1, txID:d0a2ee339923cc9739bc7900765cb29b72f0a73c7876306fadd7d5ba79a0f332
2021/02/21 15:26:25 成功发送 fabric 交易, ticketID:1, txID:d0a2ee339923cc9739bc7900765cb29b72f0a73c7876306fadd7d5ba79a0f332
...
2021/02/21 15:26:27 transaction index 0: type: ENDORSER_TRANSACTION, txid: d0a2ee339923cc9739bc7900765cb29b72f0a73c7876306fadd7d5ba79a0f332, validation code: VALID
```

### 4.完成公证人转账交易

```shell
./build/notary-cli approve --ticket-id $ticketId
```

### 5.查询跨链交易信息

```shell
./build/notary-cli query --ticket-id $ticketId
```

## 查询跨链后资产信息

```bash
# 以太账户资产信息
./build/notary-cli account --network-type ethereum --account $AliceEthAcc
./build/notary-cli account --network-type ethereum --account $BobEthAcc

# fabric资产信息
./build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $AliceFabricAcc
./build/notary-cli account --network-type fabric --fchannel mychannel --fcc basic --account $BobFabricAcc
```

