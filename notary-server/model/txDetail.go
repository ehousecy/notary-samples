package model

import (
	"database/sql"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

const TxDetailTableName = "tb_tx_detail"

type TxDetail struct {
	BaseTxDetail
	UpdateTxDetailModel
}

type BaseTxDetail struct {
	ID        int64  `db:"id" json:"id"`
	TxFrom    string `db:"tx_from" json:"from"`
	TxTo      string `db:"tx_to" json:"to"`
	Amount    string `db:"amount" json:"amount"`
	TxStatus  string `db:"tx_status" json:"tx_status"`
	Type      string `db:"type" json:"type"`
	CrossTxID int64  `db:"cross_tx_id" json:"cross_tx_id"`
}

type UpdateTxDetailModel struct {
	FromTxID       string    `db:"from_tx_id" json:"from_tx_id"`
	ToTxID         string    `db:"to_tx_id" json:"to_tx_id"`
	FromTxCreateAt time.Time `db:"from_tx_create_at" json:"from_tx_create_at"`
	ToTxCreateAt   time.Time `db:"to_tx_create_at" json:"to_tx_create_at"`
	FromTxFinishAt time.Time `db:"from_tx_finish_at" json:"from_tx_finish_at"`
	ToTxFinishAt   time.Time `db:"to_tx_finish_at" json:"to_tx_finish_at"`
}

func init() {
	InitDB()
	//创建表
	createTableSql := `
	CREATE TABLE IF NOT EXISTS ` + TxDetailTableName + `(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tx_from VARCHAR(64) NULL,
		tx_to VARCHAR(64) NULL,
		amount VARCHAR(64) NULL,
		tx_status VARCHAR(64) NULL,
		type VARCHAR(64) NULL,
		cross_tx_id INTEGER,
		from_tx_id VARCHAR(64) NULL,
		to_tx_id VARCHAR(64) NULL,
		from_tx_create_at TIMESTAMP NULL,
		to_tx_create_at timestamp NULL,
		from_tx_finish_at timestamp NULL,
		to_tx_finish_at timestamp NULL
   );`
	DB.MustExec(createTableSql)
}

func (td TxDetail) GetByCIDAndType() (*TxDetail, error) {
	return getByCTxIDAndType(td.CrossTxID, td.Type)
}

func (td TxDetail) VerifyExistedValidCIDAndType() (bool, error) {
	txDetail, err := getByCTxIDAndType(td.CrossTxID, td.Type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	//todo:添加无效状态判断
	if txDetail != nil && txDetail.TxStatus != "" {
		return true, nil
	}
	return false, nil
}

func (td TxDetail) Save(tx *sqlx.Tx) (int64, error) {
	return InsertTxDetail(td, tx)
}

func (td *TxDetail) CompleteTransferFromTx(tx *sqlx.Tx) error {
	td.TxStatus = "fromCompleted"
	td.FromTxFinishAt = time.Now()
	rows, err := UpdateTxDetailByID(*td, tx)
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("完成转账交易失败")
	}
	return nil
}

func (td *TxDetail) BoundTransferToTx(txID string, tx *sqlx.Tx) error {
	td.ToTxID = txID
	td.ToTxCreateAt = time.Now()
	td.TxStatus = "waitingToComplete"
	rows, err := UpdateTxDetailByID(*td, tx)
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("绑定交易失败")
	}
	return nil
}

func (td *TxDetail) CompleteTransferToTx(tx *sqlx.Tx) error {
	//校验需要完成的交易状态
	//todo:交易状态替换
	if td.TxStatus != "..." {
		return errors.New("当前交易不能完成")
	}
	td.TxStatus = "Completed"
	td.ToTxFinishAt = time.Now()
	rows, err := UpdateTxDetailByID(*td, tx)
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("完成转账交易失败")
	}
	return nil
}

func InsertTxDetail(td TxDetail, tx ...*sqlx.Tx) (int64, error) {
	insertMap, err := Struct2Map(td.BaseTxDetail)
	delete(insertMap, "id")

	var result sql.Result
	insertBuilder := sq.Insert(TxDetailTableName).SetMap(insertMap)
	if len(tx) > 0 {
		result, err = insertBuilder.RunWith(tx[0]).Exec()
	} else {
		result, err = insertBuilder.RunWith(DB).Exec()
	}
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func UpdateTxDetailByID(td TxDetail, tx ...*sqlx.Tx) (int64, error) {
	update, err := Struct2Map(td.UpdateTxDetailModel)
	update["tx_status"] = td.TxStatus
	updateBuilder := sq.Update(TxDetailTableName).SetMap(update).Where(sq.Eq{"id": td.ID})
	var result sql.Result
	if len(tx) > 0 {
		result, err = updateBuilder.RunWith(tx[0]).Exec()
	} else {
		result, err = updateBuilder.RunWith(DB).Exec()
	}

	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	return rows, err
}

func getByCTxIDAndType(cid int64, t string) (*TxDetail, error) {
	builder := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"cross_tx_id": cid, "type": t})
	return execGetSql(builder)
}

func getByCTxIDAndTypeAndStatus(cid int64, t string, status string) (*TxDetail, error) {
	builder := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"cross_tx_id": cid, "type": t, "tx_status": status})
	return execGetSql(builder)
}

func execGetSql(builder sq.SelectBuilder) (*TxDetail, error) {
	querySql, args, err := builder.ToSql()
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	octd := &TxDetail{}
	err = DB.Get(octd, querySql, args...)
	if err != nil {
		log.Panicln(err)
	}
	return octd, err
}

func GetTxDetailByCTxID(cid ...int64) ([]*TxDetail, error) {

	querySql, args, err := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"cross_tx_id": cid}).ToSql()
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	var octd []*TxDetail
	err = DB.Select(&octd, querySql, args...)
	if err != nil {
		log.Panicln(err)
	}
	return octd, err
}

func GetCrossTxByFromTxID(txID string, cid ...int64) (*TxDetail, error) {
	querySql, args, err := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"from_tx_id": txID}).ToSql()
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	octd := &TxDetail{}
	err = DB.Get(octd, querySql, args...)
	if err != nil {
		log.Panicln(err)
	}
	if len(cid) > 0 && octd.CrossTxID != cid[0] {
		return nil, errors.New("交易id不匹配")
	}
	return octd, err
}
