package eth

import (
	"context"
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/db/constant"
	"github.com/ehousecy/notary-samples/notary-server/db/services"
	pb "github.com/ehousecy/notary-samples/proto"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

const (
	NotaryAddress = "0x71BE5a9044F3E41c74b7c25AA19B528dd6B9f387"
)

var (
	provider = services.NewCrossTxDataServiceProvider()
)

type EthHanlder struct {
	client  *ethclient.Client //https rpc
	monitor *EthMonitor
}

// create a ethereum handler with endpoint url
func NewEthHandler(url string) *EthHanlder {
	c, err := ethclient.Dial(url)
	if err != nil {
		log.Printf("Create client failed: %v", err)
		return nil
	}
	handler := &EthHanlder{
		client:  c,
		monitor: NewMonitor("ws://localhost:3334"),
	}
	handler.loop()
	handler.monitor.Start()
	return handler
}

// loop event and confirm transaction
func (e *EthHanlder) loop() {
	events := make(chan txConfirmEvent, 100)
	e.monitor.Subscribe(events)
	go func(event chan txConfirmEvent) {
		for {
			select {
			case txEvent := <-events:
				err := e.ConfirmTx(txEvent.txid)
				if err != nil {
					EthLogPrintf("Confirm Tx error, %v", err)
				}
			default:

			}
		}
	}(events)
}

// build and sign ethereum transaction
// 3 input fields are required, they are: fromAddress, to, amount
func (e *EthHanlder) BuildTx(args ...string) *types.Transaction {
	if len(args) != 3 {
		log.Printf("Input error, should input from/to/amount/priv\n")
		return nil
	}
	// validate args and build transactions
	from := args[0]
	to := args[1]
	amount := args[2]

	// if build transaction failed, return empty bytes
	tx := e.buildTx(from, to, amount)
	return tx
}

// sign transaction and send out to the network, record transaction id locally
func (e *EthHanlder) SignAndSendTx(ticketId string, txData *types.Transaction) error {
	priv := "478976d8cfae83fdc3152c85f5c49c7c324298bc4431ee64b3caebda15fdfbfb"
	signed := e.signTx(priv, txData)
	return e.sendTx(ticketId, signed)
}

// construct and sign transactions for user
func (e *EthHanlder) ConstructAndSignTx(src pb.NotaryService_SubmitTxServer, recv *pb.TransferPropertyRequest) error {
	ticketId := recv.CTxId
	err := provider.ValidateEnableCreateTransferFromTx(ticketId, constant.TypeEthereum)
	if err != nil {
		return err
	}
	provider := services.NewCrossTxDataServiceProvider()
	info, err := provider.QueryCrossTxInfoByCID(ticketId)
	if err != nil {
		return err
	}
	from := info.EthFrom
	amount := info.EthAmount

	// build raw transaction from notary service
	rawTx := e.buildTx(from, NotaryAddress, amount)

	chainId, _ := e.client.NetworkID(context.Background())
	signer := types.NewEIP155Signer(chainId)
	h := signer.Hash(rawTx)
	err = src.Send(&pb.TransferPropertyResponse{
		Error:  nil,
		TxData: h.Bytes(),
	})
	if err != nil {
		return err
	}

	// receive signed tx from client
	signed, err := src.Recv()
	if err != nil {
		return err
	}
	signature := signed.Data
	tx, err := rawTx.WithSignature(signer, signature)
	if err != nil {
		EthLogPrintf("Sign tx failed, %v", err)
		return err
	}
	return e.sendTx(ticketId, tx)
}

// approve a cross transaction

func (e *EthHanlder) Approve(ticketId string) error {
	err := provider.ValidateEnableBoundTransferToTx(ticketId, nil)
	if err != nil {
		EthLogPrintf("validate failed: %v", err)
		return err
	}
	ticketInfo, err := provider.QueryCrossTxInfoByCID(ticketId)
	if err != nil {
		return err
	}
	rawTx := e.BuildTx(NotaryAddress, ticketInfo.EthTo, ticketInfo.EthAmount)
	err = e.SignAndSendTx(ticketId, rawTx)
	//todo:err
	return err
}

// handle confirm tx event
func (e *EthHanlder) ConfirmTx(txHash string) error {
	err := provider.CompleteTransferTx(txHash)
	return err
}

// helper functions

// sign transactions using private key
func (e *EthHanlder) signTx(priv string, tx *types.Transaction) *types.Transaction {
	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		EthLogPrintf("Invalid private key: %v\n", err)
		return nil
	}
	chainId, err := e.client.NetworkID(context.Background())
	if err != nil {
		EthLogPrintf("Failed to get chain-ID %v\n", err)
		return nil
	}
	signer := types.NewEIP155Signer(chainId)
	h := signer.Hash(tx)

	sig, err := crypto.Sign(h[:], privateKey)
	if err != nil {
		EthLogPrintf("sign tranaction failed: %v", err)
	}
	signed, err := tx.WithSignature(signer, sig)
	if err != nil {
		EthLogPrintf("append signature failed: %v", err)
		return nil
	}
	return signed
}

// send transaction
func (e *EthHanlder) sendTx(ticketId string, signed *types.Transaction) error {

	err := e.client.SendTransaction(context.Background(), signed)
	if err != nil {
		return err
	}
	txHash := signed.Hash().String()
	provider.CreateTransferFromTx(ticketId, txHash, constant.TypeEthereum)
	return e.monitor.AddTxRecord(txHash)
}

// validate private key and derive public key
func EthloggerPrint(content string) {
	log.Printf("[Eth handler] %s\n", content)
}

func EthLogPrintf(content string, v ...interface{}) {
	ss := fmt.Sprintf(content, v...)
	EthloggerPrint(ss)
}

// build tx
func (e *EthHanlder) buildTx(from, to, amount string) *types.Transaction {

	fromAddr := common2.HexToAddress(from)
	nonce, err := e.client.PendingNonceAt(context.Background(), fromAddr)
	if err != nil {
		log.Printf("Get nonce err: %v\n", err)
		return nil
	}
	gasPrice, err := e.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Printf("Get gas price failed %v\n", err)
		return nil
	}

	toAddress := common2.HexToAddress(to)

	txAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		log.Printf("Transaction amount not correct\n")
		return nil
	}

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, txAmount, 210000, gasPrice, data)
	return tx
}


