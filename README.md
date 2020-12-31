```mermaid
sequenceDiagram
	participant clientA
	participant clientB
	participant 中间人
	participant ethereum
	participant fabric
	
	#创建跨链交易单
	clientA ->> + 中间人:创建交易单：A的eth地址、转账金额，<br>A和B的fabric账号、交易的资产信息、<br>通道名称、合约名称
	中间人 ->> 中间人: 创建对应跨链交易单，状态：
	中间人 -->> -clientA:单号、中间人的eth地址
	
	#第一次以太交易：A->中间人
	clientA ->>+ 中间人: 构造转账交易并使用A的私钥签名交易--from:A,to:中间人
	中间人->> ethereum: 验证交易内容，修改交易单状态:，发送交易到ethereum
	中间人 -->>- clientA: 返回交易id,交易状态
	
	loop 监听发送的交易
		中间人 ->>+ ethereum:通过交易id查询交易
		ethereum -->>- 中间人:返回交易信息
		中间人 ->> 中间人:解析交易信息,修改交易单状态：
	end
	
	#第二次fabric交易:B->A
	clientB ->>+ 中间人:构造转账交易并使用B的私钥签名交易--from:B,to:A，eth address：B
	中间人 ->> fabric:验证交易内容，修改交易单状态:，发送交易到fabric
	中间人-->>- clientB: 交易id,单号,B的eth地址
	
	loop 监听发送的交易
		中间人 ->>+ fabric:getTransactionByID
		fabric -->>- 中间人:返回交易信息
		中间人 ->> 中间人:解析交易信息,修改交易单状态:
	end
	
	#第三次以太交易：中间人->B
	中间人->>+ ethereum: 发送转账交易，from:中间人，to：clientB
	
	loop 监听区块
		ethereum ->> 中间人:返回区块
		中间人 ->> 中间人:解析区块验证txId,修改交易单状态
	end
	
	
	
	
	
```

> todo：

- 使用语言：go|~~node~~
- 跨链交易单存储：~~文件~~|db|~~区块~~
- 版本：fabric 2.2、ethereum、sdk
- 交易限制：
  - fabric不能多次交易，不能影响非跨链交易
  - clientA取消交易限制
  - ...
- 跨链交易单状态流转：
  - created
  - escrow
  - transfer
  - settlement
  - canceled
  - ...
- 交易判断方式：块监听|通过交易id查询交易
- 以太交易校验方式：合约|普通账户







交易流程：

1.  register  Cross-chain transaction：A的eth地址、转账金额，A和B的fabric账号、交易的资产信息、通道名称、合约名称
2.  ETH-transfer：A -> 中间人
3. 更新注册收据状态：verify eth txid
4.  fabric-transfer：B -> A
5.  更新注册收据状态：verify fabric txid
6.  ETH-transfer：中间人 -> B
7.  confirm transaction completion：verify eth transaction 



> bug

- 一个ETH txid绑定多个Cross-chain register  receipt
- A同时注册两个跨链交易：仅eth转账金额不一样，B完成fabric交易后，A将txid绑定到金额较低的register  receipt



> 原因:

- register  receipt和ETH txid绑定
- register  receipt和fabric txid绑定



> 解决方案

1. 通过中间人发送交易
2. 中间人以服务的形式处理跨链操作

> 中间人服务        

中间人服务包含以下功能模块
1. create Cross-chain tx
2. list Cross-chain tx
3. monitor tx with Cross-chain txid
4. verify each transaction(from, to, amount, channel)
5. commit final transaction both side

> ```中间人服务``` 开发内容

##### client

1. 构造ETH及fabric交易并签名   2+2+2 
2. 发起跨链交易请求并返回跨链交易标识（index or cross-chain txid）1+2
3. 查询、展示跨链交易信息 1


##### cross-chain transaction service

1. 创建cross-chain tx  
2. 数据库增删改查接口
3. 交易查询修改  2
4. 交易校验（ETH && fabric） 1 + 1
5. 监听交易执行结果 0.5 + 0.5
6. 确认跨链交易，并构造发起最终确认交易（包含ETH以及Fabric）0.5 + 0.5


##### 网络搭建及模块集成

* ETH 私网搭建 0.5
    * 创建账户 
    * 配置协作账户
* Fabric 网络搭建 1.5
    * 网络部署
    * chain-code 安装
    
    
##### test && debug
 
 pending : 1 + 1  

##### 推文

2 + 2 

``` total: 24/2 = 12day ```

##### open issue

1. fabric go-sdk 需改造源码，实现构造和发送交易接口（签名，背书，发送）
2. grpc
3. rollback not considered