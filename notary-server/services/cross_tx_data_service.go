package services

import "time"

type CrossTxBase struct {
	ID              string `json:"id"`
	EthFrom         string `json:"eth_from" validate:"required"`
	EthTo           string `json:"eth_to" validate:"required"`
	EthAmount       string `json:"eth_amount" validate:"required"`
	FabricFrom      string `json:"fabric_from" validate:"required"`
	FabricTo        string `json:"fabric_to" validate:"required"`
	FabricAmount    string `json:"fabric_amount" validate:"required"`
	FabricChannel   string `json:"fabric_channel" validate:"required"`
	FabricChaincode string `json:"fabric_chaincode" validate:"required"`
}

type TxDetail struct {
	TxFrom         string    `json:"from"`
	TxTo           string    `json:"to"`
	Amount         string    `json:"amount"`
	TxStatus       string    `json:"tx_status"`
	Type           string    `json:"type"`
	CrossTxID      string    `json:"cross_tx_id"`
	FromTxID       string    `json:"from_tx_id"`
	ToTxID         string    `json:"to_tx_id"`
	FromTxCreateAt time.Time `json:"from_tx_create_at"`
	ToTxCreateAt   time.Time `json:"to_tx_create_at"`
	FromTxFinishAt time.Time `json:"from_tx_finish_at"`
	ToTxFinishAt   time.Time `json:"to_tx_finish_at"`
}

type CrossTxInfo struct {
	CrossTxBase
	FabricTx  *TxDetail
	EthTx     *TxDetail
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CrossTxDataService interface {
	CreateCrossTx(CrossTxBase) (string, error)
	CreateTransferFromTx(cid string, txID string, txType string) error
	CompleteTransferFromTx(cid string, txID string) error
	BoundTransferToTx(cid string, boundTxID, txID string) error
	CompleteTransferToTx(cid string, txID string) error
	QueryCrossTxInfoByCID(string) (*CrossTxInfo, error)
	QueryAllCrossTxInfo() ([]CrossTxInfo, error)
}
