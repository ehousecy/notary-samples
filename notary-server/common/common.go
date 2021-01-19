package common

type TxExecResult struct {
	Err error
	TxReceipt string
}

type TxHandler interface {
	ValidateTx([]byte, string) bool
	MonitorTx(txid string) chan TxExecResult
	BuildTx(args ...string) []byte
	SignTx(priv string) []byte
	ConfirmTx(txid string) bool
}