package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/ehousecy/notary-samples/notary-server/constant"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

const CrossTxDetailTableName = "tb_cross_tx_detail"

type CrossTxDetail struct {
	BaseCrossTxDetail
	UpdateCrossTxDetailModel
}

type BaseCrossTxDetail struct {
	ID              int64     `json:"id" db:"id"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	EthFrom         string    `json:"eth_from" db:"eth_from"`
	EthTo           string    `json:"eth_to" db:"eth_to"`
	EthAmount       string    `json:"eth_amount" db:"eth_amount"`
	FabricFrom      string    `json:"fabric_from" db:"fabric_from"`
	FabricTo        string    `json:"fabric_to" db:"fabric_to"`
	FabricAmount    string    `json:"fabric_amount" db:"fabric_amount"`
	FabricChannel   string    `json:"fabric_channel" db:"fabric_channel"`
	FabricChaincode string    `json:"fabric_chaincode" db:"fabric_chaincode"`
}

type UpdateCrossTxDetailModel struct {
	EthToNotaryAt      time.Time `json:"eth_to_notary_at" db:"eth_to_notary_at"`
	FabricToNotaryAt   time.Time `json:"fabric_to_notary_at" db:"fabric_to_notary_at"`
	EthFromNotaryAt    time.Time `json:"eth_from_notary_at" db:"eth_from_notary_at"`
	FabricFromNotaryAt time.Time `json:"fabric_from_notary_at" db:"fabric_from_notary_at"`
	FinishedAt         time.Time `json:"finished_at" db:"finished_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	EthToTxID          string    `json:"eth_to_tx_id" db:"eth_to_tx_id"`
	FabricToTxID       string    `json:"fabric_to_tx_id" db:"fabric_to_tx_id"`
	EthFromTxID        string    `json:"eth_from_tx_id" db:"eth_from_tx_id"`
	FabricFromTxID     string    `json:"fabric_from_tx_id" db:"fabric_from_tx_id"`
}

type originalCrossTxDetail struct {
	BaseCrossTxDetail
	originalUpdateCrossTxData
}

type originalUpdateCrossTxData struct {
	EthToNotaryAt      sql.NullTime   `json:"eth_to_notary_at" db:"eth_to_notary_at"`
	FabricToNotaryAt   sql.NullTime   `json:"fabric_to_notary_at" db:"fabric_to_notary_at"`
	EthFromNotaryAt    sql.NullTime   `json:"eth_from_notary_at" db:"eth_from_notary_at"`
	FabricFromNotaryAt sql.NullTime   `json:"fabric_from_notary_at" db:"fabric_from_notary_at"`
	FinishedAt         sql.NullTime   `json:"finished_at" db:"finished_at"`
	UpdatedAt          sql.NullTime   `json:"updated_at" db:"updated_at"`
	EthToTxID          sql.NullString `json:"eth_to_tx_id" db:"eth_to_tx_id"`
	FabricToTxID       sql.NullString `json:"fabric_to_tx_id" db:"fabric_to_tx_id"`
	EthFromTxID        sql.NullString `json:"eth_from_tx_id" db:"eth_from_tx_id"`
	FabricFromTxID     sql.NullString `json:"fabric_from_tx_id" db:"fabric_from_tx_id"`
}

func init() {
	InitDB()
	//创建表
	createTableSql := ` 
	CREATE TABLE IF NOT EXISTS ` + CrossTxDetailTableName + `(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		eth_from VARCHAR(64) NULL,
		eth_to VARCHAR(64) NULL,
		eth_amount VARCHAR(64) NULL,
		fabric_from VARCHAR(64) NULL,
		fabric_to VARCHAR(64) NULL,
		fabric_amount VARCHAR(64) NULL,
		fabric_channel VARCHAR(64) NULL,
		fabric_chaincode VARCHAR(64) NULL,
		status VARCHAR(64) NULL,
		created_at TIMESTAMP NULL,
		eth_to_notary_at timestamp NULL,
		fabric_to_notary_at timestamp NULL,
		eth_from_notary_at timestamp NULL,
		fabric_from_notary_at timestamp NULL,
		finished_at timestamp NULL,
		updated_at timestamp NULL,
		eth_to_tx_id VARCHAR(64) NULL,
		fabric_to_tx_id VARCHAR(64) NULL,
		eth_from_tx_id VARCHAR(64) NULL,
		fabric_from_tx_id VARCHAR(64) default ''
    );`
	DB.MustExec(createTableSql)
}

func (ctd CrossTxDetail) Save() (int64, error) {
	return InsertCrossTxDetail(ctd)
}

func (ctd CrossTxDetail) GetById() (*CrossTxDetail, error) {
	return GetCrossTxDetailByID(ctd.ID)
}

func (ctd CrossTxDetail) Update(tx ...*sqlx.Tx) error {
	row, err := UpdateCrossTxDetailByID(ctd, tx...)
	if err != nil {
		return err
	}
	if row != 1 {
		err = errors.New("更新失败")
	}
	return err
}

func (ctd *CrossTxDetail) CreateTransferFromTxInfo(txID string, t string, tx ...*sqlx.Tx) error {
	switch t {
	case constant.TypeEthereum:
		ctd.EthFromTxID = txID
	case constant.TypeFabric:
		ctd.FabricFromTxID = txID
	default:
		return errors.New("交易类型不支持")
	}
	ctd.UpdatedAt = time.Now()
	err := ctd.Update(tx...)
	if err != nil {
		return err
	}
	return nil
}

func (ctd *CrossTxDetail) CompleteTransferFromTx(txID string, t string, tx ...*sqlx.Tx) error {
	switch t {
	case constant.TypeEthereum:
		if ctd.EthFromTxID != txID {
			return errors.New("完成转账交易失败，txID不匹配")
		}
		ctd.EthFromNotaryAt = time.Now()
	case constant.TypeFabric:
		if ctd.FabricFromTxID != txID {
			return errors.New("完成转账交易失败，txID不匹配")
		}
		ctd.FabricFromNotaryAt = time.Now()
	default:
		return errors.New("交易类型不支持")
	}
	ctd.UpdatedAt = time.Now()
	err := ctd.Update(tx...)
	if err != nil {
		return err
	}
	return nil
}

func (ctd *CrossTxDetail) BoundTransferToTxInfo(txID string, t string, tx ...*sqlx.Tx) error {
	switch t {
	case constant.TypeEthereum:
		ctd.EthToTxID = txID
	case constant.TypeFabric:
		ctd.FabricToTxID = txID
	default:
		return errors.New("交易类型不支持")
	}
	//todo:检查是否两个交易托管交易都完成，完成更新跨链交易状态

	ctd.UpdatedAt = time.Now()
	err := ctd.Update(tx...)
	if err != nil {
		return err
	}
	return nil
}

func (ctd *CrossTxDetail) CompleteTransferToTx(txID string, t string, tx ...*sqlx.Tx) error {
	switch t {
	case constant.TypeEthereum:
		if ctd.EthToTxID != txID {
			return errors.New("完成转账交易失败，txID不匹配")
		}
		ctd.EthToNotaryAt = time.Now()
	case constant.TypeFabric:
		if ctd.FabricToTxID != txID {
			return errors.New("完成转账交易失败，txID不匹配")
		}
		ctd.FabricToNotaryAt = time.Now()
	default:
		return errors.New("交易类型不支持")
	}
	//todo:检查是否两个分发交易都完成 完成修改最终状态

	ctd.UpdatedAt = time.Now()
	err := ctd.Update(tx...)
	if err != nil {
		return err
	}
	return nil
}

func InsertCrossTxDetail(ctd CrossTxDetail) (int64, error) {
	ctd.CreatedAt = time.Now()
	insertMap, err := Struct2Map(ctd.BaseCrossTxDetail)
	delete(insertMap, "id")

	result, err := sq.Insert(CrossTxDetailTableName).SetMap(insertMap).RunWith(DB).Exec()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func UpdateCrossTxDetailByID(ctd CrossTxDetail, tx ...*sqlx.Tx) (int64, error) {
	ctd.UpdatedAt = time.Now()
	updateMap, err := Struct2Map(ctd.UpdateCrossTxDetailModel)
	updateMap["status"] = ctd.Status
	var result sql.Result
	updateBuilder := sq.Update(CrossTxDetailTableName).SetMap(updateMap).Where(sq.Eq{"id": ctd.ID})
	if tx != nil && len(tx) > 0 {
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

func GetCrossTxDetailByID(id int64) (*CrossTxDetail, error) {
	querySql, args, err := sq.Select("*").From(CrossTxDetailTableName).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	octd := &originalCrossTxDetail{}
	err = DB.Get(octd, querySql, args...)
	if err != nil {
		log.Panicln(err)
	}
	return octd.convert(), err
}

func GetCrossTxDetails() ([]*CrossTxDetail, error) {
	var octd []*originalCrossTxDetail
	querySql, _, err := sq.Select("*").From(CrossTxDetailTableName).ToSql()
	if err != nil {
		return nil, err
	}
	err = DB.Select(&octd, querySql)
	if err != nil {
		return nil, err
	}
	return convertOriginalArr(octd), err

}

func convertOriginalArr(octdArr []*originalCrossTxDetail) []*CrossTxDetail {
	var ctdArr []*CrossTxDetail
	for _, octd := range octdArr {
		ctdArr = append(ctdArr, octd.convert())
	}
	return ctdArr
}

func (octd *originalCrossTxDetail) convert() *CrossTxDetail {
	var uctd UpdateCrossTxDetailModel
	data := octd.originalUpdateCrossTxData
	if data.EthToNotaryAt.Valid {
		uctd.EthToNotaryAt = data.EthToNotaryAt.Time
	}
	if data.FabricToNotaryAt.Valid {
		uctd.FabricToNotaryAt = data.FabricToNotaryAt.Time
	}
	if data.EthFromNotaryAt.Valid {
		uctd.EthFromNotaryAt = data.EthFromNotaryAt.Time
	}
	if data.FabricFromNotaryAt.Valid {
		uctd.FabricFromNotaryAt = data.FabricFromNotaryAt.Time
	}
	if data.FinishedAt.Valid {
		uctd.FinishedAt = data.FinishedAt.Time
	}
	if data.UpdatedAt.Valid {
		uctd.UpdatedAt = data.UpdatedAt.Time
	}
	if data.EthToTxID.Valid {
		uctd.EthToTxID = data.EthToTxID.String
	}
	if data.FabricToTxID.Valid {
		uctd.FabricToTxID = data.FabricToTxID.String
	}
	if data.EthFromTxID.Valid {
		uctd.EthFromTxID = data.EthFromTxID.String
	}
	if data.FabricFromTxID.Valid {
		uctd.FabricFromTxID = data.FabricFromTxID.String
	}

	return &CrossTxDetail{
		//ID:                       octd.ID,
		BaseCrossTxDetail:        octd.BaseCrossTxDetail,
		UpdateCrossTxDetailModel: uctd,
	}
}

func Struct2Map(s interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	return m, err
}

func String(v interface{}) {
	bytes, _ := json.Marshal(v)
	fmt.Println(string(bytes))
}
