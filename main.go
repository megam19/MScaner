package main

import (
	"MScaner/packages"
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

var timeSleep = 60 //в секундах
var arr_items_DBStruct []itemStruct
var arr_items_FolderStruct []itemStruct
var i_folder itemStruct

type itemStruct struct {
	fileName string
	fileSize int64
}

func main() {
	fmt.Println("Список файлов в: " + DIR_PATH)
	packages.ConnectToDB()

	for {
		arr_items_DBStruct = arr_items_DBStruct[:0]         //обнуление слайса но остается прежняя capacity
		arr_items_FolderStruct = arr_items_FolderStruct[:0] //обнуление слайса но остается прежняя capacity
		readDatabase()
		scanDir(DIR_PATH)

		//Слайс для хранения информации о новых файлах, т.е. разницы между базой и папкой
		items := GetItemsDifferents(arr_items_DBStruct, arr_items_FolderStruct)

		for _, item := range items {
			fmt.Println("Новый файл: " + item.fileName)
			packages.WriteDatabase(item.fileName, item.fileSize) //Запись в базу данных
		}

		time.Sleep(time.Duration(timeSleep) * time.Second) // Повторять каждые timeSleep секунд
	}
}

// Функция возвращает элементы, которые:
// - есть в folder, но нет в db   → новые
// - есть в обоих, но размер отличается → изменённые
func GetItemsDifferents(db []itemStruct, folder []itemStruct) []itemStruct {
	// Карта для быстрого поиска: имя файла → размер в базе
	dbMap := make(map[string]int64, len(db))

	for _, item := range db {
		dbMap[item.fileName] = item.fileSize
	}

	var results []itemStruct

	// Проходим по текущему состоянию папки
	for _, f := range folder {
		if dbSize, exists := dbMap[f.fileName]; !exists || dbSize != f.fileSize {
			// либо файла вообще нет в базе
			// либо размер изменился
			results = append(results, f)
		}
	}

	return results
}

func scanDir(dirPath string) {

	files, err := os.ReadDir(dirPath) // Расчет на 20-40к файлов
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".mxf" {
			fmt.Println("Совсем не MXF: " + file.Name())
			continue
		}

		info, err := file.Info()
		if err != nil {
			log.Printf("Не удалось получить инфо о файле %s: %v", file.Name(), err)
		}

		size := info.Size() / 1024 //Преобразование в КБ

		i_folder = itemStruct{file.Name(), size}                          //наполняем структуру
		arr_items_FolderStruct = append(arr_items_FolderStruct, i_folder) // наполняем массив структурами
	}

}

func readDatabase() {
	var i_db itemStruct

	packages.Database, packages.ErrDB = sql.Open("sqlite", packages.DatabasePath)
	if packages.ErrDB != nil {
		log.Fatal("База данных недоступен " + packages.ErrDB.Error())
	}
	defer packages.Database.Close()

	querySELECT := `SELECT fileName, fileSize FROM files;`
	rows, errDB := packages.Database.Query(querySELECT)
	if errDB != nil {
		log.Fatalf("Ошибка запроса в базу данных")
	}
	defer rows.Close() // сразу после проверки ошибки

	for rows.Next() {
		err := rows.Scan(&i_db.fileName, &i_db.fileSize) // наполняем структуру
		if err != nil {
			fmt.Println(err)
			continue
		}

		arr_items_DBStruct = append(arr_items_DBStruct, i_db) // наполняем массив структурами
	}

	fmt.Println("******************* Конец чтения Базы **********************")

}
