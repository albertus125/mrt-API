package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}
	if err = DB.Ping(); err != nil {
		return err
	}
	log.Println("Database connection established")
	return nil

}

func GetDB() *sql.DB {
	return DB
}
