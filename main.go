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
const timeSleep = 60 //в минутах
const deletePeriodDays = 25

func main() {
	fmt.Println("Список файлов в: " + DIR_PATH)
	dbConnect := packages.DB_connect()

	triggers.AutoPurgeFilesAndDB(dbConnect, deletePeriodDays, DIR_PATH) //Будет создан триггер на каждый день 00:00

	for {

		Arr_FilesInfo := packages.ScanDir(DIR_PATH)
		Arr_items_DBStruct := packages.ReadDb(dbConnect)

		//Слайс для хранения информации о новых файлах, т.е. разницы между базой и папкой
		items := packages.DifferentsToWriteDB(Arr_items_DBStruct, Arr_FilesInfo)

		for _, item := range items {
			fmt.Println("Найден новый файл: " + item.FileName)
			packages.WriteToDB(dbConnect, item.FileName, item.FileSize) //Запись в базу данных
		}

		log.Println("Следующее через", timeSleep, "минут...")
		log.Println("")
		time.Sleep(time.Duration(timeSleep) * time.Minute) // Повторять каждые timeSleep секунд
	}
}
