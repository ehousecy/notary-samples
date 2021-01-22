package model

import (
	"github.com/jmoiron/sqlx"
	//_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

var DB *sqlx.DB
var once sync.Once

func init() {
	InitDB()
}

func InitDB() {
	once.Do(func() {
		DB = sqlx.MustConnect("sqlite3", "./foo.db")
	})
}
