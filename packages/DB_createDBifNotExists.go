package packages

import "log"

func CreateDBifNotExists() {

	queryCreateDB := `CREATE TABLE IF NOT EXISTS files(
						id INTEGER PRIMARY KEY AUTOINCREMENT,	
						fileName TEXT NOT NULL UNIQUE, 
						fileSize INTEGER, 
						createdAt DATETIME DEFAULT CURRENT_TIMESTAMP, 
						updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
					)`
	queryPrepare, err := Database.Prepare(queryCreateDB)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %q", err)
	}
	queryPrepare.Exec()
}
