package main

import (
	"MScaner/packages"
	"MScaner/triggers"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite" // Важно: имя драйвера "sqlite"
)

const DIR_PATH = "\\\\air-02\\imagine_mxf"

//const DIR_PATH = "\\\\fserver\\harris"

const timeSleep = 120 //в секундах

func main() {
	fmt.Println("Список файлов в: " + DIR_PATH)
	dbConnect := packages.DB_connect()

	triggers.AutoPurgeFilesAndDB(dbConnect, 20, DIR_PATH) //Будет создан триггер на каждый день 00:00

	for {

		packages.DB_read(dbConnect)
		packages.ScanDir(DIR_PATH)

		//Слайс для хранения информации о новых файлах, т.е. разницы между базой и папкой
		items := packages.DifferentsToWriteDB(packages.Arr_items_DBStruct, packages.Arr_items_FolderStruct)

		for _, item := range items {
			fmt.Println("Найден новый файл: " + item.FileName)
			packages.DB_write(item.FileName, item.FileSize) //Запись в базу данных
		}

		log.Println("Сканирование завершено. Следующее через", timeSleep, "секунд...")
		time.Sleep(time.Duration(timeSleep) * time.Second) // Повторять каждые timeSleep секунд
	}
}
