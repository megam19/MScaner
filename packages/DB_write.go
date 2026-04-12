package packages

import (
	"database/sql"
	"log"
)

func WriteDatabase(fileName string, fileSize int64) {
	DB, _ = sql.Open("sqlite", DatabasePath)
	//defer DB.Close()

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
