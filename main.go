package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite" // Важно: имя драйвера "sqlite"
)

// var DIR_PATH = "\\\\10.33.6.202\\na_prosmotr"
// var DIR_PATH = "\\\\10.33.6.202\\lowres\\LOWRES_ARCHIVE"
// var DIR_PATH = "\\\\fserver\\harris"
var DIR_PATH = "\\\\air-02\\imagine_mxf"

var database *sql.DB
var timeSpeep = 60 //в секундах

func main() {
	var errdb error
	fmt.Println("Список файлов в: " + DIR_PATH)
	database, errdb = sql.Open("sqlite", "./database/sqlite3DB.db")
	if errdb != nil {
		log.Fatalf("Ошибка открытия базы %d", errdb)
	}
	defer database.Close()

	queryCreateDB := `CREATE TABLE IF NOT EXISTS files(
						id INTEGER PRIMARY KEY AUTOINCREMENT,	
						fileName TEXT NOT NULL UNIQUE, 
						fileSize INTEGER, 
						createdAt DATETIME DEFAULT CURRENT_TIMESTAMP, 
						updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
					)`
	statement, err := database.Prepare(queryCreateDB)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %q", err)
	}
	statement.Exec()

	for {
		scanDir(DIR_PATH)
		time.Sleep(time.Duration(timeSpeep) * time.Second)
	}
}

func scanDir(dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		info, _ := file.Info()
		size := info.Size() / 1024 //Преобразование в КБ
		fSize := strconv.Itoa(int(size))

		fmt.Println(file.Name() + " Размер: " + fSize + " КБ") // выременный вывод в консоль
		writeDatabase(file.Name(), size)
	}
}

func writeDatabase(fileName string, fileSize int64) {
	database, _ = sql.Open("sqlite", "./database/sqlite3DB.db")
	defer database.Close()

	//Запрос записывает в случае отсутствует fileName или изменился размер файла.
	queryINSERT := `INSERT INTO files (fileName, fileSize) VALUES (?, ?)
			  ON CONFLICT(fileName) DO UPDATE SET 
				fileSize = excluded.fileSize,
				updatedAt = CURRENT_TIMESTAMP;`

	result, _ := database.Prepare(queryINSERT)
	result.Exec(fileName, fileSize)
}
