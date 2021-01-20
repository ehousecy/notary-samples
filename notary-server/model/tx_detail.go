package model

import (
	"database/sql"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/ehousecy/notary-samples/notary-server/constant"
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
	TxFrom    string `db:"tx_from" json:"tx_from"`
	TxTo      string `db:"tx_to" json:"tx_to"`
	Amount    string `db:"amount" json:"amount"`
	TxStatus  string `db:"tx_status" json:"tx_status"`
	Type      string `db:"type" json:"type"`
	CrossTxID int64  `db:"cross_tx_id" json:"cross_tx_id"`
	ChannelID string `db:"channel_id" json:"channel_id"`
}

type UpdateTxDetailModel struct {
	FromTxID       string    `db:"from_tx_id" json:"from_tx_id"`
	ToTxID         string    `db:"to_tx_id" json:"to_tx_id"`
	FromTxCreateAt time.Time `db:"from_tx_create_at" json:"from_tx_create_at"`
	ToTxCreateAt   time.Time `db:"to_tx_create_at" json:"to_tx_create_at"`
	FromTxFinishAt time.Time `db:"from_tx_finish_at" json:"from_tx_finish_at"`
	ToTxFinishAt   time.Time `db:"to_tx_finish_at" json:"to_tx_finish_at"`
}

type originalTxDetail struct {
	BaseTxDetail
	originalUpdateTxDetailModel
}
type originalUpdateTxDetailModel struct {
	FromTxID       string         `db:"from_tx_id" json:"from_tx_id"`
	ToTxID         sql.NullString `db:"to_tx_id" json:"to_tx_id"`
	FromTxCreateAt time.Time      `db:"from_tx_create_at" json:"from_tx_create_at"`
	ToTxCreateAt   sql.NullTime   `db:"to_tx_create_at" json:"to_tx_create_at"`
	FromTxFinishAt sql.NullTime   `db:"from_tx_finish_at" json:"from_tx_finish_at"`
	ToTxFinishAt   sql.NullTime   `db:"to_tx_finish_at" json:"to_tx_finish_at"`
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
		channel_id VARCHAR(64) '',
		from_tx_id VARCHAR(64) NULL,
		to_tx_id VARCHAR(64) NULL,
		from_tx_create_at TIMESTAMP NULL,
		to_tx_create_at timestamp NULL,
		from_tx_finish_at timestamp NULL,
		to_tx_finish_at timestamp NULL
   );`
	DB.MustExec(createTableSql)
}

func (td TxDetail) GetByCIDAndType() ([]*TxDetail, error) {
	return getByCTxIDAndType(td.CrossTxID, td.Type)
}

func ValidateExistedValidTxDetailCIDAndType(cid int64, txType string) (bool, error) {
	//查询是否存在
	txDetails, err := getByCTxIDAndType(cid, txType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	//判断是否无效
	//todo:添加无效状态判断
	if len(txDetails) > 0 {
		for _, td := range txDetails {
			if td.TxStatus != constant.TxStatusFromFinished {
				return true, nil
			}
		}
	}
	return false, nil
}

func NewTransferFromTx(ctd *CrossTxDetail, txType string, txID string) TxDetail {
	amount := ctd.FabricAmount
	txFrom := ctd.FabricFrom
	if txType == constant.TypeEthereum {
		amount = ctd.EthAmount
		txFrom = ctd.EthFrom
	}
	td := TxDetail{
		BaseTxDetail: BaseTxDetail{
			TxFrom:    txFrom,
			Amount:    amount,
			TxStatus:  constant.TxStatusFromCreated,
			Type:      txType,
			CrossTxID: ctd.ID,
			ChannelID: ctd.FabricChannel,
		},
		UpdateTxDetailModel: UpdateTxDetailModel{
			FromTxID:       txID,
			FromTxCreateAt: time.Now(),
		},
	}
	return td
}
func (td TxDetail) Save(tx *sqlx.Tx) (int64, error) {
	td.TxStatus = constant.TxStatusFromCreated
	td.FromTxCreateAt = time.Now()
	return insertTxDetail(td, tx)
}

func (td *TxDetail) CompleteTransferFromTx(tx *sqlx.Tx) error {
	td.TxStatus = constant.TxStatusFromFinished
	td.FromTxFinishAt = time.Now()
	rows, err := updateTxDetailByID(*td, tx)
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("完成转账交易失败")
	}
	return nil
}

func (td *TxDetail) BoundTransferToTx(txID string, tx *sqlx.Tx) error {
	if td.TxStatus != constant.TxStatusFromFinished {
		return errors.New("当前交易不能代理转账")
	}
	td.ToTxID = txID
	td.ToTxCreateAt = time.Now()
	td.TxStatus = constant.TxStatusToCreated
	rows, err := updateTxDetailByID(*td, tx)
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
	if td.TxStatus != constant.TxStatusToCreated {
		return errors.New("当前交易不能完成")
	}
	td.TxStatus = constant.TxStatusToFinished
	td.ToTxFinishAt = time.Now()
	rows, err := updateTxDetailByID(*td, tx)
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("完成转账交易失败")
	}
	return nil
}

func insertTxDetail(td TxDetail, tx ...*sqlx.Tx) (int64, error) {
	insertMap, err := Struct2Map(td.BaseTxDetail)
	delete(insertMap, "id")
	insertMap["from_tx_id"] = td.FromTxID
	insertMap["from_tx_create_at"] = td.FromTxCreateAt

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

func updateTxDetailByID(td TxDetail, tx ...*sqlx.Tx) (int64, error) {
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

func getByCTxIDAndType(cid int64, t string) ([]*TxDetail, error) {
	builder := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"cross_tx_id": cid, "type": t})
	return execSelectSql(builder)
}

func getByCTxIDAndTypeAndStatus(cid int64, t string, status string) (*TxDetail, error) {
	builder := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"cross_tx_id": cid, "type": t, "tx_status": status})
	return execGetSql(builder, cid)
}

func execGetSql(builder sq.SelectBuilder, cid ...int64) (*TxDetail, error) {
	querySql, args, err := builder.ToSql()
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	td := &originalTxDetail{}
	err = DB.Get(td, querySql, args...)
	if err != nil {
		log.Panicln(err)
	}
	if len(cid) > 0 && td.CrossTxID != cid[0] {
		return nil, errors.New("交易id不匹配")
	}
	return td.convert(), err
}

func execSelectSql(builder sq.SelectBuilder) ([]*TxDetail, error) {
	querySql, args, err := builder.ToSql()
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	var otd []*originalTxDetail
	err = DB.Select(&otd, querySql, args...)
	if err != nil {
		log.Println(err)
	}
	return convertOriginalTxDetailArr(otd), err
}

func GetTxDetailByCTxID(cid ...int64) ([]*TxDetail, error) {
	builder := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"cross_tx_id": cid})
	return execSelectSql(builder)
}

func GetTxDetailByFromTxID(txID string, cid ...int64) (*TxDetail, error) {
	getSql := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"from_tx_id": txID})
	return execGetSql(getSql, cid...)
}

func GetTxDetailByToTxID(txID string, cid ...int64) (*TxDetail, error) {
	getSql := sq.Select("*").From(TxDetailTableName).Where(sq.Eq{"to_tx_id": txID})
	return execGetSql(getSql, cid...)
}

func GetConfirmingTxDetailByType(txType string) ([]*TxDetail, error) {
	and := sq.And{sq.Eq{"type": txType},
		sq.Or{sq.Eq{"tx_status": constant.TxStatusToCreated}, sq.Eq{"tx_status": constant.TxStatusFromCreated}}}
	builder := sq.Select("*").From(TxDetailTableName).Where(and)
	return execSelectSql(builder)
}

func (otd *originalTxDetail) convert() *TxDetail {
	utdm := UpdateTxDetailModel{}
	data := otd.originalUpdateTxDetailModel
	if data.ToTxID.Valid {
		utdm.ToTxID = data.ToTxID.String
	}
	if data.FromTxFinishAt.Valid {
		utdm.FromTxFinishAt = data.FromTxFinishAt.Time
	}
	if data.ToTxCreateAt.Valid {
		utdm.ToTxCreateAt = data.ToTxCreateAt.Time
	}
	if data.ToTxFinishAt.Valid {
		utdm.ToTxFinishAt = data.ToTxFinishAt.Time
	}
	utdm.FromTxID = data.FromTxID
	utdm.FromTxCreateAt = data.FromTxCreateAt
	return &TxDetail{
		BaseTxDetail:        otd.BaseTxDetail,
		UpdateTxDetailModel: utdm,
	}
}
func convertOriginalTxDetailArr(otdArr []*originalTxDetail) []*TxDetail {
	var tdArr []*TxDetail
	for _, oct := range otdArr {
		tdArr = append(tdArr, oct.convert())
	}
	return tdArr
}
