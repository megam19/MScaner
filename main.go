package main

import (
	"MScaner/packages"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Важно: имя драйвера "sqlite"
)

// var DIR_PATH = "\\\\10.33.6.202\\na_prosmotr"
// var DIR_PATH = "\\\\10.33.6.202\\lowres\\LOWRES_ARCHIVE"
//const DIR_PATH = "\\\\air-02\\imagine_mxf"

const DIR_PATH = "\\\\fserver\\harris"

const timeSleep = 60 //в секундах

func main() {
	fmt.Println("Список файлов в: " + DIR_PATH)
	dbConnect := packages.ConnectToDB()

	for {

		packages.ReadDatabase(dbConnect)
		packages.ScanDir(DIR_PATH)

		//Слайс для хранения информации о новых файлах, т.е. разницы между базой и папкой
		items := packages.GetItemsDifferents(packages.Arr_items_DBStruct, packages.Arr_items_FolderStruct)

		for _, item := range items {
			fmt.Println("Новый файл: " + item.FileName)
			packages.WriteDatabase(item.FileName, item.FileSize) //Запись в базу данных
		}
		fmt.Println("Цикл завершена")

		time.Sleep(time.Duration(timeSleep) * time.Second) // Повторять каждые timeSleep секунд
	}
}
