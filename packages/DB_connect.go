package packages

import (
	"database/sql"
	"log"
)

var DB *sql.DB
var queryPrepare *sql.Stmt

const DatabasePath = "./database/sqlite3DB.db"

func DB_connect() *sql.DB {

	DB, err := sql.Open("sqlite", DatabasePath)
	if err != nil {
		log.Fatalf("Ошибка открытия базы %d", err)
	}

	queryCreateDB := `CREATE TABLE IF NOT EXISTS files(
						fileName TEXT NOT NULL UNIQUE, 
						fileSize INTEGER, 
						createdAt DATETIME DEFAULT CURRENT_TIMESTAMP, 
						updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
					)`
	queryPrepare, err := DB.Prepare(queryCreateDB)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %q", err)
	}
	queryPrepare.Exec()

	return DB
}
