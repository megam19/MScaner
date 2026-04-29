package triggers

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type item_db struct {
	fileName  string
	fileSize  int64
	updatedAt string
}

func AutoPurgeFilesAndDB(db *sql.DB, deletePeriodDays int, dirPath string) {

	var count = 0

	timenow := time.Now().UTC() //Берем текущую дату в UTC
	rows, err := db.Query(`SELECT fileName, fileSize, updatedAt FROM files;`)
	if err != nil {
		log.Println("Ошибка запроса в базу данных")
	}
	defer rows.Close()

	for rows.Next() {
		var item item_db
		err := rows.Scan(&item.fileName, &item.fileSize, &item.updatedAt)
		if err != nil {
			log.Println(err)
			continue
		}

		//Сравнивает дату если старше указанных дней deletePeriodDay, удаляем
		if item.updatedAt <= timenow.AddDate(0, 0, -deletePeriodDays).Format("2006-01-02T15:04:05Z") { //Специальный формат для Go "2006-01-02T15:04:05Z"
			count++
			fmt.Println("На удаление: ", item.updatedAt+" "+item.fileName)

			//deleteInDB("2018_03_31_WFCA_46_1_3.mxf", db)
		}
	}
	fmt.Println("Количество найденных файлов:", count)
}

func deleteInDB(id string, db *sql.DB) {

	query := `DELETE FROM files WHERE fileName=$1`
	res, err := db.Exec(query, id)
	if err != nil {
		log.Printf("ошибка выполнения запроса для файла %s: %v", id, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Println("Не удалось получить RowsAffected %w", err)
		return
	}
	fmt.Printf("Удалено файлов: %d", rowsAffected)
}
