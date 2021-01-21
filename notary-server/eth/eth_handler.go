package eth

import (
	"context"
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
	"math/big"
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


// add validate rules here
func (e *EthHanlder) ValidateTx(signedBytes []byte, ticketId string) bool {
	signedTx := decodeTx(signedBytes)
	_ = signedTx
	//todo
	// get ticket data from db and validate amounts

	return true
}

// monitorTx should monitor transaction execute result, return error if transaction failed.
func (e *EthHanlder) SendTx(txData []byte)  {

	// monitor logic here
	return
}

// build and sign ethereum transaction
func (e *EthHanlder) BuildTx(args ...string) []byte {
	if len(args) != 3 {
		log.Printf("Input error, should input from/to/amount/priv\n")
		return []byte{}
	}
	// validate args and build transactions
	priv := args[0]
	to := args[1]
	amount := args[2]

	pub, err := getPublicAddr(priv)
	if err != nil {
		log.Printf("Invalid private key: %v\n", err)
		return []byte{}
	}
	// if build transaction failed, return empty bytes
	tx := e.buildTx(pub, to, amount)
	if tx == nil {
		log.Printf("Build transaction failed!\n")
		return []byte{}
	}

	// sign transaction
	signed := e.signTx(priv, tx)
	if signed == nil {
		log.Printf("Sign transaction failed\n")
		return []byte{}
	}

	// using rlp encode transaction to bytes
	signedBytes, err := rlp.EncodeToBytes(signed)
	if err != nil {
		log.Printf("Encode transaction failed: %v\n", err)
		return []byte{}
	}
	return signedBytes
}

func (e *EthHanlder)SignTx(priv string, ticketId string) []byte  {
	return []byte{}
}

// helper functions

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
func getPublicAddr(priv string) (string, error) {
	privKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		return "", err
	}
	pubAddress := crypto.PubkeyToAddress(privKey.PublicKey)
	return pubAddress.String(), nil
}

func EthloggerPrint(content string) {
	log.Printf("[Eth handler] %s\n", content)
}

func EthLogPrintf(content string, v ...interface{}) {
	ss := fmt.Sprintf(content, v...)
	EthloggerPrint(ss)
}
