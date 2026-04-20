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

var database_c *sql.DB

func AutoPurgeFilesAndDB(db *sql.DB, deletePeriodDays int, dirPath string) {

	var count = 0
	database_c = db
	timenow := time.Now().UTC() //Берем текущую дату в UTC
	rows, err := db.Query(`SELECT fileName, fileSize, updatedAt FROM files;`)
	if err != nil {
		log.Println("Ошибка запроса в базу данных")
	}
	//defer rows.Close()

	for rows.Next() {
		var item item_db
		err := rows.Scan(&item.fileName, &item.fileSize, &item.updatedAt) // наполняем структуру
		if err != nil {
			log.Println(err)
			continue
		}
		//Сравнивает дату если старше указанных дней deletePeriodDay, удаляем
		if item.updatedAt <= timenow.AddDate(0, 0, -deletePeriodDays).Format("2006-01-02T15:04:05Z") { //Специальный формат для Go "2006-01-02T15:04:05Z"
			count++
			fmt.Println(item.updatedAt + " " + item.fileName)

			//time.Sleep(time.Duration(1) * time.Second)
			//deleteInDB(item.fileName, db)
		}
	}
	fmt.Println("Количество найденных файлов:", count)
}

func deleteInDB(id string) {

	query := `DELETE FROM files WHERE fileName="$1"`
	res, err := database_c.Exec(query, id)
	if err != nil {
		log.Printf("ошибка выполнения запроса для файла %s: %v", id, err)
	}
	rowsAffected, _ := res.RowsAffected()
	fmt.Printf("Удалено файлов: %d", rowsAffected)
}
