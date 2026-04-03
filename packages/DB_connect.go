package packages

import (
	"database/sql"
	"log"
)

var Database *sql.DB
var queryPrepare *sql.Stmt
var DatabasePath = "./database/sqlite3DB.db"
var ErrDB error

func ConnectToDB() {

	Database, ErrDB = sql.Open("sqlite", DatabasePath)
	if ErrDB != nil {
		log.Fatalf("Ошибка открытия базы %d", ErrDB)
	}
	defer Database.Close()

}
