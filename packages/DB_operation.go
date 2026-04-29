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

func DB_read(db *sql.DB) []ItemStruct {

	var i_db ItemStruct
	var Arr_items_DBStruct []ItemStruct

	querySELECT := `SELECT fileName, fileSize FROM files;`
	rows, err := db.Query(querySELECT)
	if err != nil {
		log.Println("Ошибка запроса в базу данных")
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&i_db.FileName, &i_db.FileSize) // наполняем структуру
		if err != nil {
			log.Println(err)
			continue
		}

		Arr_items_DBStruct = append(Arr_items_DBStruct, i_db) // наполняем массив структурами
	}

	log.Println("Завершена чтение базы данных")
	return Arr_items_DBStruct
}

func DB_write(DB *sql.DB, fileName string, fileSize int64) {

	//Запрос записывает в случае отсутствует fileName или изменился размер файла.
	queryINSERT := `INSERT INTO files (fileName, fileSize) VALUES (?, ?)
			  ON CONFLICT(fileName) DO UPDATE SET 
				fileSize = excluded.fileSize,
				updatedAt = CURRENT_TIMESTAMP;`

	result, err := DB.Prepare(queryINSERT)
	if err != nil {
		log.Fatal("Ошибка в записи в базу данных")
	}
	defer result.Close()

	_, err = result.Exec(fileName, fileSize)
	if err != nil {
		log.Printf("ошибка выполнения запроса для файла %s: %v", fileName, err)
	}

	log.Printf("✓ Файл записан/обновлён: %s (%d KB)", fileName, fileSize)
}
