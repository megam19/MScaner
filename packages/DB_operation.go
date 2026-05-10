package packages

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

//var queryPrepare *sql.Stmt

const DatabasePath = "./database/sqlite3DB.db"
const connStr = "host=192.168.109.205 port=5432 user=postgres password=aA123456 dbname=mscaner sslmode=disable"

func DB_connect() *sql.DB {
	//Провайдеры: sqlite, postgres
	DB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка открытия базы %v", err)
	}
	//SQLITE
	/*queryCreateDB := `CREATE TABLE IF NOT EXISTS files(
		fileName TEXT NOT NULL UNIQUE,
		fileSize INTEGER,
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
	)`*/

	//PostgreSQL
	queryCreateDB := `CREATE TABLE IF NOT EXISTS files (
						fileName TEXT NOT NULL UNIQUE,
						fileSize BIGINT,
						createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);`
	_, err = DB.Exec(queryCreateDB)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %q", err)
	}

	return DB
}

func DB_read(db *sql.DB) []ItemStruct {

	var i_db ItemStruct
	var Arr_items_DBStruct []ItemStruct

	querySELECT := `SELECT fileName, fileSize FROM files;`
	rows, err := db.Query(querySELECT)
	if err != nil {
		log.Println("Ошибка запроса в базу данных", err.Error())
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

func DB_write(db *sql.DB, fileName string, fileSize int64) {

	//Запрос записывает в случае отсутствует fileName или изменился размер файла.
	queryINSERT := `INSERT INTO files (fileName, fileSize) VALUES ($1, $2)
			  ON CONFLICT(fileName) DO UPDATE SET 
				fileSize = excluded.fileSize,
				updatedAt = CURRENT_TIMESTAMP;`

	_, err := db.Exec(queryINSERT, fileName, fileSize)
	if err != nil {
		log.Printf("Ошибка в записи в базу данных %s: %v", fileName, err.Error())
	}

	log.Printf("✓ Файл записан/обновлён: %s (%d KB)", fileName, fileSize)
}
func DB_Delete(db *sql.DB, fileName string) {
	test := "2018_03_31_WFCA_46_1_3.mxf"

	_, err := db.Exec("DELETE FROM files WHERE fileName=$1", test)
	if err != nil {
		fmt.Println("не удалось очистить с базы", test, err.Error())
	}
	//fmt.Println("С базы удален", fileName)
}
