package packages

import (
	"database/sql"
	"log"
)

var Database *sql.DB
var queryPrepare *sql.Stmt

const DatabasePath = "./database/sqlite3DB.db"

func ConnectToDB() *sql.DB {

	Database1, err := sql.Open("sqlite", DatabasePath)
	if err != nil {
		log.Fatalf("Ошибка открытия базы %d", err)
	}
	//defer Database.Close()
	return Database1
}
