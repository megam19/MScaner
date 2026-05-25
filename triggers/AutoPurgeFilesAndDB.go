package triggers

import (
	"MScaner/packages"
	"database/sql"
	"fmt"
	"log"
	"time"
)

func AutoPurgeFilesAndDB(db *sql.DB, deletePeriodDays int, dirPath string) {

	porogDney := time.Now().UTC().AddDate(0, 0, -deletePeriodDays) //Берем текущую дату в UTC с минусом дней deletePeriodDays
	deletedCount := 0

	rows, err := db.Query(`
		SELECT fileName, fileSize, updatedAt 
		FROM files 
		WHERE updatedAt < $1;
		`, porogDney)
	if err != nil {
		log.Printf("Ошибка запроса к БД %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item struct {
			fileName  string
			fileSize  int64
			updatedAt time.Time
		}

		err := rows.Scan(&item.fileName, &item.fileSize, &item.updatedAt)
		if err != nil {
			log.Printf("Ошибка сканировния строки: %v", err)
			continue
		}
		fmt.Printf("На удаление: %s (%s)\n", item.fileName, item.updatedAt.Format(time.RFC3339))

		packages.DeleteInDB(db, item.fileName)
		deletedCount++
	}
	log.Printf("✅ AutoPurge завершён, удалено записей: %d", deletedCount)
	fmt.Println("")
}
