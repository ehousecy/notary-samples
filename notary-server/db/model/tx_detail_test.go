package model

import (
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/db/constant"
	"github.com/jmoiron/sqlx"
	"testing"
)

var cid int64 = 3
var fromTxID = fmt.Sprintf("12456-%v", cid)
var toTxID = fmt.Sprintf("456789-%v", cid)

func TestTxDetail_Save(t *testing.T) {
	ctd, err := GetCrossTxDetailByID(cid)
	if err != nil {
		t.Fatal(err)
	}
	txDetail := NewTransferFromTx(ctd, constant.TypeFabric, fromTxID)
	tx := DB.MustBegin()
	defer rollbackTx(tx, t)
	save, err := txDetail.Save(tx)
	if err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	t.Logf("insert tx detail id=%v", save)

}

func TestTxDetail_CompleteTransferFromTx(t *testing.T) {
	td, err := GetTxDetailByFromTxID(fromTxID)
	if err != nil {
		t.Fatal(err)
	}
	tx := DB.MustBegin()
	defer rollbackTx(tx, t)
	if err = td.CompleteTransferFromTx(tx); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	t.Logf("success Complete Transfer From Tx,fromTxID=%v", fromTxID)
}

func TestTxDetail_BoundTransferToTx(t *testing.T) {
	td, err := GetTxDetailByFromTxID(fromTxID)
	if err != nil {
		t.Fatal(err)
	}
	tx := DB.MustBegin()
	defer rollbackTx(tx, t)
	if err = td.BoundTransferToTx(toTxID, tx); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	t.Logf("success Complete Transfer From Tx, fromTxID=%v, toTxID=%v", fromTxID, toTxID)
}

func TestTxDetail_CompleteTransferToTx(t *testing.T) {

}

func TestGetConfirmingTxDetailByType(t *testing.T) {
	txDetails, err := GetConfirmingTxDetailByType(constant.TypeFabric)
	if err != nil {
		t.Fatal(err)
	}
	for _, detail := range txDetails {
		fmt.Println(detail)
	}
}

func rollbackTx(tx *sqlx.Tx, t *testing.T) {
	if err := tx.Rollback(); err != nil {
		t.Logf("unable to rollback: %v", err)
	}
}
