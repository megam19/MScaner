package packages

import (
	"database/sql"
	"log"
)

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
