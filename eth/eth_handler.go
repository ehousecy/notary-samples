package eth

import (
	"github.com/ehousecy/notary-samples/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthHanlder struct {
	client ethclient.Client
}

func (e *EthHanlder)ValidateTx(signedTx []byte) bool {

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