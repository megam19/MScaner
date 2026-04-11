package packages

import (
	"database/sql"
	"fmt"
	"log"
)

func ReadDatabase(db *sql.DB) {

	Arr_items_DBStruct = Arr_items_DBStruct[:0] //обнуление слайса но остается прежняя capacity
	var i_db ItemStruct

	querySELECT := `SELECT fileName, fileSize FROM files;`
	rows, err := db.Query(querySELECT)
	if err != nil {
		log.Fatalf("Ошибка запроса в базу данных")
	}
	defer rows.Close() // сразу после проверки ошибки

	for rows.Next() {
		err := rows.Scan(&i_db.FileName, &i_db.FileSize) // наполняем структуру
		if err != nil {
			fmt.Println(err)
			continue
		}

		Arr_items_DBStruct = append(Arr_items_DBStruct, i_db) // наполняем массив структурами
	}

	fmt.Println("******* Конец чтения Базы *********")

}
