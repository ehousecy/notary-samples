package services

import (
	"encoding/json"
	"github.com/ehousecy/notary-samples/notary-server/db/constant"
	"github.com/stretchr/testify/require"
	"testing"
)

var cid = "1"
var provider = NewCrossTxDataServiceProvider()
var fabricFromTxID = "ff:tx123-" + cid
var ethFromTxID = "ef:tx123-" + cid
var fabricToTxID = "ff:tx456-" + cid
var ethToTxID = "ef:tx456-" + cid

type txParam struct {
	cid       string
	txID      string
	txType    string
	boundTxID string
}

// go test -timeout 30s github.com/ehousecy/notary-samples/notary-server/db/services -run ^TestCrossTxDataServiceProvider_CreateCrossTx_Success$ -v
func TestCrossTxDataServiceProvider_CreateCrossTx_Success(t *testing.T) {
	ctb := CrossTxBase{
		EthFrom:         "0X123456",
		EthTo:           "0X789456",
		EthAmount:       "10",
		FabricFrom:      "Bill",
		FabricTo:        "Gary",
		FabricAmount:    "asset1",
		FabricChannel:   "mychannel",
		FabricChaincode: "basic",
	}
	id, err := provider.CreateCrossTx(ctb)
	require.NoErrorf(t, err, "failed create cross tx, err=%v", err)
	t.Logf("success create cross tx cid=%v", id)
	cid = id
	//验证状态
	cti := getCrossTxInfoByCID(cid, t)
	require.Equal(t, constant.StatusCreated, cti.Status, "failed validate cross tx status")
}

// go test -timeout 30s github.com/ehousecy/notary-samples/notary-server/db/services -run ^TestCrossTxDataServiceProvider_CreateTransferFromTx_Success$ -v
func TestCrossTxDataServiceProvider_CreateTransferFromTx_Success(t *testing.T) {
	var txTests = []txParam{
		{cid: cid, txID: fabricFromTxID, txType: constant.TypeFabric},
		{cid: cid, txID: ethFromTxID, txType: constant.TypeEthereum},
	}
	for _, tt := range txTests {
		if err := provider.CreateTransferFromTx(tt.cid, tt.txID, tt.txType); err != nil {
			t.Fatalf("failed create %v Transfer From Tx, err=%v", tt.txType, err)
		}
		//验证状态
		cti := getCrossTxInfoByCID(tt.cid, t)
		tx := getTxDetail(tt.txType, cti)
		require.Equal(t, constant.TxStatusFromCreated, tx.TxStatus, "failed validate %v Transfer From Tx status", tt.txType)
	}
}

// go test -timeout 30s github.com/ehousecy/notary-samples/notary-server/db/services -run ^TestCrossTxDataServiceProvider_CompleteTransferFromTx_Success$ -v
func TestCrossTxDataServiceProvider_CompleteTransferFromTx_Success(t *testing.T) {
	var txTests = []txParam{
		{cid: cid, txID: fabricFromTxID, txType: constant.TypeFabric},
		{cid: cid, txID: ethFromTxID, txType: constant.TypeEthereum},
	}

	for _, tt := range txTests {
		err := provider.ValidateEnableCompleteTransferTx(tt.txID)
		require.NoError(t, err, "should no err validate enable complete Transfer from tx")
		t.Logf("this cross tx enable complete %v Transfer From Tx, cid=%v, txID=%v", tt.txType, tt.cid, tt.txID)
		if err := provider.CompleteTransferTx(tt.txID); err != nil {
			t.Fatalf("failed complete %v Transfer From Tx, err=%v", tt.txType, err)
		}
		//校验状态
		cti := getCrossTxInfoByCID(tt.cid, t)
		tx := getTxDetail(tt.txType, cti)
		require.Equal(t, constant.TxStatusFromFinished, tx.TxStatus, "failed validate %v Transfer From Tx status", tt.txType)
	}
	cti := getCrossTxInfoByCID(txTests[0].cid, t)
	require.Equal(t, constant.StatusHosted, cti.Status, "failed validate cross Tx status")
}

// go test -timeout 30s github.com/ehousecy/notary-samples/notary-server/db/services -run ^TestCrossTxDataServiceProvider_BoundTransferToTx_Success$ -v
func TestCrossTxDataServiceProvider_BoundTransferToTx_Success(t *testing.T) {
	var txTests = []txParam{
		{cid: cid, txID: fabricToTxID, txType: constant.TypeFabric, boundTxID: fabricFromTxID},
		{cid: cid, txID: ethToTxID, txType: constant.TypeEthereum, boundTxID: ethFromTxID},
	}
	for _, tt := range txTests {
		err := provider.ValidateEnableBoundTransferToTx(tt.boundTxID, nil)
		require.NoError(t, err, "failed validate %v enable bound transfer to tx", tt.txType)
		err = provider.BoundTransferToTx(tt.boundTxID, tt.txID)
		require.NoError(t, err, "failed bound transfer to tx, cid=%s, boundID=%s, txID=%s", tt.cid, tt.boundTxID, tt.txID)
		td := getTxDetail(tt.txType, getCrossTxInfoByCID(tt.cid, t))
		require.Equal(t, constant.TxStatusToCreated, td.TxStatus, "failed validate %v bound transfer to tx status", tt.txType)
	}
}

// go test -timeout 30s github.com/ehousecy/notary-samples/notary-server/db/services -run ^TestCrossTxDataServiceProvider_CompleteTransferToTx_Success$ -v
func TestCrossTxDataServiceProvider_CompleteTransferToTx_Success(t *testing.T) {
	var txTests = []txParam{
		{cid: cid, txID: fabricToTxID, txType: constant.TypeFabric},
		{cid: cid, txID: ethToTxID, txType: constant.TypeEthereum},
	}
	for _, tt := range txTests {
		err := provider.ValidateEnableCompleteTransferTx(tt.txID)
		require.NoError(t, err, "failed validate %v enable complete transfer to tx", tt.txType)
		err = provider.CompleteTransferTx(tt.txID)
		require.NoError(t, err, "failed complete transfer to tx, cid=%s, txID=%s", tt.cid, tt.txID)
		td := getTxDetail(tt.txType, getCrossTxInfoByCID(tt.cid, t))
		require.Equal(t, constant.TxStatusToFinished, td.TxStatus, "failed validate %v complete transfer to tx status", tt.txType)
	}
	cti := getCrossTxInfoByCID(txTests[0].cid, t)
	require.Equal(t, constant.StatusFinished, cti.Status, "failed validate cross Tx status")
}

// go test -timeout 30s github.com/ehousecy/notary-samples/notary-server/db/services -run ^TestCrossTxDataServiceProvider_QueryConfirmingTxInfo$ -v
func TestCrossTxDataServiceProvider_QueryConfirmingTxInfo(t *testing.T) {
	typeTests := []string{constant.TypeEthereum, constant.TypeFabric}
	for _, tt := range typeTests {
		confirmingTxInfos, err := provider.QueryConfirmingTxInfo(tt)
		if err != nil {
			t.Fatalf("failed query %v confirming tx list, err=%v", tt, err)
		}
		t.Logf("success query %v confirming tx list, list=%v", tt, confirmingTxInfos)
	}
}

// go test -timeout 30s github.com/ehousecy/notary-samples/notary-server/db/services -run ^TestCrossTxDataServiceProvider_QueryCrossTxInfoByCID$ -v
func TestCrossTxDataServiceProvider_QueryCrossTxInfoByCID(t *testing.T) {
	cti, err := provider.QueryCrossTxInfoByCID(cid)
	require.NoErrorf(t, err, "failed query cross tx info, cid=%v, err=%v", cid, err)
	marshal, err := json.Marshal(cti)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cross tx info: %v", string(marshal))
}

func TestCrossTxDataServiceProvider_CancelTransferTx(t *testing.T) {
	//provider.CancelTransferTx(ethFromTxID)
	provider.CancelTransferTx(ethToTxID)
}

func TestCrossTxDataServiceProvider_FailTransferTx(t *testing.T) {
	//provider.FailTransferTx(ethFromTxID)
	provider.FailTransferTx(ethToTxID)
}

func getCrossTxInfoByCID(cid string, t *testing.T) *CrossTxInfo {
	cti, err := provider.QueryCrossTxInfoByCID(cid)
	require.NoErrorf(t, err, "failed query cross tx info, cid=%v, err=%v", cid, err)
	return cti
}
func getTxDetail(txType string, cti *CrossTxInfo) *TxDetail {
	var tx *TxDetail
	if txType == constant.TypeFabric {
		tx = cti.FabricTx
	} else {
		tx = cti.EthTx
	}
	return tx
}
