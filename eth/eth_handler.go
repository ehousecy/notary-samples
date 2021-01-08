package eth

import (
	"github.com/ehousecy/notary-samples/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

type EthHanlder struct {
	client ethclient.Client
}

func (e *EthHanlder)ValidateTx(signedBytes []byte, ticketId string) bool {
	// todo
	// test decode raw data
	// validate signature, to, amount
	signedTx := decodeTx(signedBytes)
	_ = signedTx


	return true
}

func (e *EthHanlder)MonitorTx(txId string) chan common.TxExecResult {
	ch := make(chan common.TxExecResult, 1)
	// todo
	// monitor logic here
	return ch
}

func (e *EthHanlder)BuildTx(args ...string) []byte {
	_ = args
	// validate args and build transactions
	rawTx := make([]byte, len(args))
	return rawTx
}

func decodeTx(tx []byte) *types.Transaction {
	var decoded *types.Transaction
	if err := rlp.DecodeBytes(tx, &decoded); err != nil {
		return nil
	}
	return decoded
}
