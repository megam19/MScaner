package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // Важно: имя драйвера "sqlite"
)

// var DIR_PATH = "\\\\10.33.6.202\\na_prosmotr"
// var DIR_PATH = "\\\\10.33.6.202\\lowres\\LOWRES_ARCHIVE"
// var DIR_PATH = "\\\\air-02\\imagine_mxf"
var DIR_PATH = "\\\\fserver\\harris"

var database *sql.DB
var queryPrepare *sql.Stmt
var databasePath = "./database/sqlite3DB.db"
var timeSpeep = 160 //в секундах
var errDB error
var arr_items_DBStruct []itemDBStruct
var arr_items_FolderStruct []itemFolderStruct

var i_folder itemFolderStruct

type itemDBStruct struct {
	fileName string
	fileSize int64
}
type itemFolderStruct struct {
	fileName string
	fileSize int64
}

func main() {
	fmt.Println("Список файлов в: " + DIR_PATH)
	createDBifNotExists()

	for {
		//readDatabase()
		scanDir(DIR_PATH)
		time.Sleep(time.Duration(timeSpeep) * time.Second)
	}
}

func createDBifNotExists() {

	database, errDB = sql.Open("sqlite", databasePath)
	if errDB != nil {
		log.Fatalf("Ошибка открытия базы %d", errDB)
	}
	defer database.Close()

	queryCreateDB := `CREATE TABLE IF NOT EXISTS files(
						id INTEGER PRIMARY KEY AUTOINCREMENT,	
						fileName TEXT NOT NULL UNIQUE, 
						fileSize INTEGER, 
						createdAt DATETIME DEFAULT CURRENT_TIMESTAMP, 
						updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
					)`
	queryPrepare, err := database.Prepare(queryCreateDB)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %q", err)
	}
	queryPrepare.Exec()
}

func differenStucts() {

}

func scanDir(dirPath string) {

	files, err := os.ReadDir(dirPath) // Расчет на 20-40к файлов
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if file.IsDir() && filepath.Ext(file.Name()) != ".mxf" {
			continue
		}

		info, _ := file.Info()
		size := info.Size() / 1024 //Преобразование в КБ

		if filepath.Ext(file.Name()) != ".mxf" { // Если папка и не mxf
			fmt.Println("Совсем не MXF: " + file.Name())
			continue
		}

		fmt.Println("MXF: " + file.Name())

		i_folder = itemFolderStruct{file.Name(), size}                    //наполняем структуру
		arr_items_FolderStruct = append(arr_items_FolderStruct, i_folder) // наполняем массив структурами

		//writeDatabase(file.Name(), size)

	}

}

func readDatabase() {
	var i_db itemDBStruct

	database, errDB = sql.Open("sqlite", databasePath)
	if errDB != nil {
		log.Fatal("База данных недоступен " + errDB.Error())
	}
	defer database.Close()

	querySELECT := `SELECT fileName, fileSize FROM files;`
	rows, errDB := database.Query(querySELECT)
	if errDB != nil {
		log.Fatalf("Ошибка запроса в базу данных")
	}

	for rows.Next() {
		err := rows.Scan(&i_db.fileName, &i_db.fileSize) // наполняем структуру
		if err != nil {
			fmt.Println(err)
			continue
		}

		arr_items_DBStruct = append(arr_items_DBStruct, i_db) // наполняем массив структурами
	}

	fmt.Printf("******************* Конец чтения Базы **********************")

}

func writeDatabase(fileName string, fileSize int64) {
	database, _ = sql.Open("sqlite", databasePath)
	defer database.Close()

	//Запрос записывает в случае отсутствует fileName или изменился размер файла.
	queryINSERT := `INSERT INTO files (fileName, fileSize) VALUES (?, ?)
			  ON CONFLICT(fileName) DO UPDATE SET 
				fileSize = excluded.fileSize,
				updatedAt = CURRENT_TIMESTAMP;`

	result, _ := database.Prepare(queryINSERT)
	result.Exec(fileName, fileSize)
}
