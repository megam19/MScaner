package packages

import (
	"database/sql"
	"log"
)

func DB_read(db *sql.DB) {

	Arr_items_DBStruct = Arr_items_DBStruct[:0] //обнуление слайса но остается прежняя capacity
	var i_db ItemStruct

	querySELECT := `SELECT fileName, fileSize FROM files;`
	rows, err := db.Query(querySELECT)
	if err != nil {
		log.Println("Ошибка запроса в базу данных")
	}
	defer rows.Close() // сразу после проверки ошибки

	for rows.Next() {
		err := rows.Scan(&i_db.FileName, &i_db.FileSize) // наполняем структуру
		if err != nil {
			log.Println(err)
			continue
		}

		Arr_items_DBStruct = append(Arr_items_DBStruct, i_db) // наполняем массив структурами
	}

	log.Println("******* Конец чтения Базы *********")

}
