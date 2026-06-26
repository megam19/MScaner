package packages

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

const connStr = "host=192.168.109.205 port=5432 user=postgres password=aA123456 dbname=mscaner sslmode=disable"

func DB_connect() *sql.DB {
	//Провайдер: postgres
	DB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка открытия базы %v", err)
	}

	//PostgreSQL
	queryCreateDB := `CREATE TABLE IF NOT EXISTS files (
						filename TEXT PRIMARY KEY,
						filesize BIGINT,
						createdat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updatedat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);`
	_, err = DB.Exec(queryCreateDB)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %q", err)
	}
	return DB
}

func ReadDb(db *sql.DB) []ItemStruct {

	var i_db ItemStruct
	var Arr_items_DBStruct []ItemStruct

	querySELECT := `SELECT filename, filesize FROM files;`
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

func WriteToDB(db *sql.DB, fileName string, fileSize int64) {

	//Запрос записывает в случае отсутствует fileName или изменился размер файла.
	queryINSERT := `INSERT INTO files (filename, filesize) VALUES ($1, $2)
					ON CONFLICT(filename) DO UPDATE SET 
					filesize = excluded.filesize,
					updatedat = CURRENT_TIMESTAMP;`

	_, err := db.Exec(queryINSERT, fileName, fileSize)
	if err != nil {
		log.Printf("Ошибка в записи в базу данных %s: %v", fileName, err.Error())
	}

	log.Printf("✓ Файл записан/обновлён: %s (%d KB)", fileName, fileSize)
}

func DeleteInDB(db *sql.DB, fileName string) {

	_, err := db.Exec("DELETE FROM files WHERE filename=$1", fileName)
	if err != nil {
		fmt.Println("не удалось очистить с базы", fileName, err.Error())
	}
}
