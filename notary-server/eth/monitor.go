package eth

import (
	"context"
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
)

// ethereum monitor
// monitor is used to listen network block events and record transaction states
// ie. pending, in block, 6 block confirmed
// db info, key: txhash, value: big.int default value 0
var  (
	numberConfirmation = new(big.Int).SetUint64(6)
) 
type EthMonitor struct {
	Client *ethclient.Client // ws rpc
	DBInterface *leveldb.DB
	done chan interface{}
}

func NewMonitor(url string) *EthMonitor {
	c, err := ethclient.Dial(url)
	if err != nil {
		return nil
	}
	db, err := leveldb.OpenFile("~/.ehouse/etherTx.db", nil)
	if err != nil {
		return nil
	}
	return &EthMonitor{
		Client: c,
		DBInterface: db,
		done: make(chan interface{}, 1),
	}
}

func (m *EthMonitor)Start()  {
	go m.loop()
}

func (m *EthMonitor)Stop()  {
	close(m.done)
}

func (m *EthMonitor)loop()  {
	headers := make(chan *types.Header)
	sub, err := m.Client.SubscribeNewHead(context.Background(), headers)
	defer close(headers)
	defer sub.Unsubscribe()
	if err != nil {
		EthLogPrintf("Failed to subscribe new header event, %v", err)
		return
	}
	for {
		select {
		case subErr := <-sub.Err():
			// try reconnect
			EthLogPrintf("Failed to subscribe new header event, %v", subErr)
			sub, err = m.Client.SubscribeNewHead(context.Background(), headers)
		case newHeader := <- headers :
			currNumber := newHeader.Number
			m.scanBlock(currNumber)
			shouldConfirmed := new(big.Int)
			shouldConfirmed.Sub(currNumber, numberConfirmation)
			m.confirmBlock(shouldConfirmed)
			// exit if stops monitor
		case <-m.done:
			return
		}
	}
}

// if a new transaction is sent out, add the transaction hash in the watching list
// called by ethereum handler
func (m *EthMonitor)AddTxRecord(txhash string) error {
	key := []byte(txhash)
	val := new(big.Int).SetUint64(0)
	err := m.DBInterface.Put(key, val.Bytes(), nil)
	return err
}


// once received a new block, scan block transactions, update the db transaction height
func (m *EthMonitor) scanBlock(blockHeight *big.Int) error {
	block, err  := m.Client.BlockByNumber(context.Background(),blockHeight)
	if err != nil {
		return err
	}
	txs := block.Transactions()
	for _, tx := range txs {
		hash := fmt.Sprintf("%s", tx.Hash())
		err := m.updateTx(hash, blockHeight.Bytes())
		if err != nil {
			EthLogPrintf("update Tx error: %v", err)
		}
	}
	EthLogPrintf("processed new block height: %s", blockHeight.String())
	return nil
}


// update transaction height record
func (m *EthMonitor)updateTx(txHash string, value []byte) error {
	key := []byte(txHash)
	exist, err := m.DBInterface.Has(key, nil)
	if err != nil {
		return err
	}
	if !exist{
		return nil
	}
	err = m.DBInterface.Put(key, value, nil)
	return err
}


// confirm all the transaction that is confirmed before the target block height
func (m *EthMonitor)confirmBlock(targetHeight *big.Int)  {
	iter := m.DBInterface.NewIterator(nil, nil)
	txHeight := new(big.Int)
	for iter.Next() {
		val := iter.Value()
		txHeight.SetBytes(val)
		if txHeight.Cmp(targetHeight) <= 0{
			txHash := fmt.Sprintf("%x", iter.Key())
			m.validateReceipt(txHash, targetHeight)
		}
	}
}

// check if confirmed

func isConfirmed(txReceipt *types.Receipt, targetHeight *big.Int) bool  {
	txHeight := txReceipt.BlockNumber
	if txHeight.Cmp(targetHeight) > 0{
		return false
	}
	return true
}

func isSuccess(txReceipt *types.Receipt) bool {
	if txReceipt.Status != types.ReceiptStatusSuccessful{
		return false
	}
	return true
}

// check transaction receipt status code, return true if tx is successfully executed in block
func (m *EthMonitor) validateReceipt(txHash string, targetHeight *big.Int)   {
	tx := common2.HexToHash(txHash)
	txReceipt,err := m.Client.TransactionReceipt(context.Background(), tx)
	if err != nil {
		EthLogPrintf("query tx Receipt failed, %v", err)
		return 
	}
	// in case txReceipt not found
	if txReceipt == nil {
		return 
	}
	
	// if transaction is not confirmed, update tx record to db
	if !isConfirmed(txReceipt, targetHeight){
		m.updateTx(txHash, txReceipt.BlockNumber.Bytes())
		return
	}
	receiptRes := isSuccess(txReceipt)
	m.confirmTx(txHash, receiptRes)
	
}

// if the transaction is confirmed with 6 blocks, delete record and notify ETH handler
func (m *EthMonitor)confirmTx(txhash string, isSuccess bool) error {
	//todo
	//notify eth handler
	key := []byte(txhash)
	err := m.DBInterface.Delete(key, nil)
	return err
}
