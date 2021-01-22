package eth

import (
	"context"
	"errors"
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/constant"
	"github.com/ehousecy/notary-samples/notary-server/services"
	pb "github.com/ehousecy/notary-samples/proto"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
	"math/big"
)

const (
	NotaryAddress = "0x71BE5a9044F3E41c74b7c25AA19B528dd6B9f387"
)

type EthHanlder struct {
	client  *ethclient.Client //https rpc
	monitor *EthMonitor
	// todo configure monitor from file
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
		monitor: NewMonitor(url),
	}
	handler.monitor.Start()
	handler.loop()
	return handler
}

func (e *EthHanlder)loop()  {
	events := make(chan txConfirmEvent, 100)
	e.monitor.Subscribe(events)
	go func (event chan txConfirmEvent){
		for{
			select {
			case txEvent := <- events:
				//
				_ = txEvent
			}
		}
	}(events)
}

// build and sign ethereum transaction
// 3 input fields are required, they are: fromAddress, to, amount
func (e *EthHanlder) BuildTx(args ...string) ([]byte) {
	if len(args) != 3 {
		log.Printf("Input error, should input from/to/amount/priv\n")
		return []byte{}
	}
	// validate args and build transactions
	from := args[0]
	to := args[1]
	amount := args[2]

	// if build transaction failed, return empty bytes
	tx := e.buildTx(from, to, amount)
	if tx == nil {
		log.Printf("Build transaction failed!\n")
		return []byte{}
	}

	// using rlp encode transaction to bytes
	rawTxBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		log.Printf("Encode transaction failed: %v\n", err)
		return []byte{}
	}
	return rawTxBytes
}

// sign transaction and send out to the network, record transaction id locally
func (e *EthHanlder)SignAndSendTx(rawData []byte) error  {
	priv := "478976d8cfae83fdc3152c85f5c49c7c324298bc4431ee64b3caebda15fdfbfb"
	privKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		EthLogPrintf("Invalid private key, %s", priv)
		return err
	}
	var tx *types.Transaction
	err = rlp.DecodeBytes(rawData, tx)
	if err != nil {
		EthloggerPrint("Invalid transaction data")
		return err
	}

	chainId, err := e.client.NetworkID(context.Background())
	if err != nil {
		EthLogPrintf("Failed to get chainId %v", err)
		return err
	}
	// sign transaction
	signed, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privKey)
	if err != nil {
		EthLogPrintf("Failed sign transaction, %v", err)
		return err
	}
	err = e.client.SendTransaction(context.Background(), signed)
	if err != nil {
		return err
	}

	//  record transaction id in monitor
	return e.monitor.AddTxRecord(signed.Hash().String())
}

// construct and sign transactions for user
func (e *EthHanlder)ConstructAndSignTx(src pb.NotaryService_SubmitTxServer) error{
	recv, err := src.Recv()
	if err != nil{
		return err
	}

	ticketId := recv.CTxId
	provider := services.NewCrossTxDataServiceProvider()
	info, err := provider.QueryCrossTxInfoByCID(ticketId)
	if err != nil {
		return err
	}
	e.BuildTx()
	from := info.EthFrom
	amount := info.EthAmount

	// build raw transaction from notary service
	rawTx := e.BuildTx(from, NotaryAddress, amount)
	if len(rawTx) ==0 {
		return errors.New("build tx failed")
	}

	err = src.Send(&pb.TransferPropertyResponse{
		Error: nil,
		TxData: rawTx,
	})
	if err != nil {
		return err
	}

	// receive signed tx from client
	signed, err := src.Recv()
	signedTx := signed.Data

	if !validateWithOrign(rawTx, signedTx){
		txSent := decodeTx(signedTx)
		err = e.client.SendTransaction(context.Background(), txSent)
		if err != nil {
			return err
		}
		txHash := txSent.Hash().String()
		provider.CreateTransferFromTx(ticketId, txHash, constant.TypeEthereum)
		return e.monitor.AddTxRecord(txHash)
	}else {
		return errors.New("transaction does not Match! ")
	}
}

// approve a cross transaction

func (e *EthHanlder)Approve(ticketId string) error  {
	provider := services.NewCrossTxDataServiceProvider()
	ticketInfo, err := provider.QueryCrossTxInfoByCID(ticketId)
	if err != nil {
		return err
	}

	rawTx := e.BuildTx(NotaryAddress, ticketInfo.EthTo, ticketInfo.EthAmount)
	err = e.SignAndSendTx(rawTx)
	return nil
}

// helper functions

func validateWithOrign(rawTx, signedTx []byte) bool {
	raw := decodeTx(rawTx)
	signed := decodeTx(signedTx)
	if raw.Hash().String() != signed.Hash().String(){
		return false
	}
	return true
}

// decode transaction from bytes
func decodeTx(tx []byte) *types.Transaction {
	var decoded *types.Transaction
	if err := rlp.DecodeBytes(tx, &decoded); err != nil {
		return nil
	}
	return decoded
}

// construct ethereum transaction base on from/to/amount/
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

// sign transactions using private key
func (e *EthHanlder) signTx(priv string, tx *types.Transaction) *types.Transaction {
	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		log.Printf("Invalid private key: %v\n", err)
		return nil
	}
	chainId, err := e.client.NetworkID(context.Background())
	if err != nil {
		log.Printf("Failed to get chain-ID %v\n", err)
		return nil
	}
	signed, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)

	if err != nil {
		log.Printf("Sign transaction failed %v\n", err)
		return nil
	}
	return signed
}

// validate private key and derive public key
func EthloggerPrint(content string) {
	log.Printf("[Eth handler] %s\n", content)
}

func EthLogPrintf(content string, v ...interface{}) {
	ss := fmt.Sprintf(content, v...)
	EthloggerPrint(ss)
}
