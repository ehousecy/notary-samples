package model

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"log"
	"time"
)

const FabricBlockLogTableName = "fabric_block_log"

type FabricBlockLog struct {
	ID          int64     `db:"id" json:"id"`
	BlockNumber uint64    `db:"block_number" json:"block_number"`
	ChannelID   string    `db:"channel_id" json:"channel_id"`
	HandleTime  time.Time `db:"handle_time" json:"handle_time"`
}

func init() {
	InitDB()
	//创建表
	createTableSql := `
	CREATE TABLE IF NOT EXISTS ` + FabricBlockLogTableName + `(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		block_number INTEGER,
		channel_id VARCHAR(64) NULL,
		handle_time TIMESTAMP NULL
   );`
	DB.MustExec(createTableSql)
}

func InsertFabricBlockLog(blockNumber uint64, channelID string) {
	blockLog := FabricBlockLog{
		BlockNumber: blockNumber,
		ChannelID:   channelID,
		HandleTime:  time.Now(),
	}
	insertMap, err := Struct2Map(blockLog)
	delete(insertMap, "id")

	result, err := sq.Insert(FabricBlockLogTableName).SetMap(insertMap).RunWith(DB).Exec()
	if err != nil {
		log.Panicf("add fabric block log record fail, err=%v", err)
		return
	}
	_, err = result.LastInsertId()
	if err != nil {
		log.Panicf("add fabric block log record fail, err=%v", err)
		return
	}
	log.Printf("add fabric block log record success, blockNumber=%v, channelID=%v", blockNumber, channelID)
}

func QueryLastFabricBlockNumber(channelID string) (uint64, error) {
	qSql, args, err := sq.Select("max(block_number)").From(FabricBlockLogTableName).Where(sq.Eq{"channel_id": channelID}).ToSql()
	if err != nil {
		return 0, err
	}
	var blockNumber sql.NullInt64
	err = DB.Get(&blockNumber, qSql, args...)
	var num = uint64(blockNumber.Int64)
	return num, err
}
