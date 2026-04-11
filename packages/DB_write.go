package packages

import (
	"database/sql"
	"fmt"
	"log"
)

func WriteDatabase(fileName string, fileSize int64) {
	Database, _ = sql.Open("sqlite", DatabasePath)
	defer Database.Close()

	//Запрос записывает в случае отсутствует fileName или изменился размер файла.
	queryINSERT := `INSERT INTO files (fileName, fileSize) VALUES (?, ?);`

	result, errDB := Database.Prepare(queryINSERT)
	if errDB != nil {
		log.Fatal("Ошибка в записи в базу данных")
	}
	fmt.Println(fileSize)

	result.Exec(fileName, fileSize)
}
