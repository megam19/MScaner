package main

import (
	"MScaner/packages"
	"MScaner/triggers"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "modernc.org/sqlite" // Важно: имя драйвера "sqlite"
)

const DIR_PATH = "\\\\air-02\\imagine_mxf"
const timeSleep = 60        //в минутах
const deletePeriodDays = 25 //дней

var purgeMutex sync.Mutex // Mutex

func main() {
	fmt.Println("Список файлов в: " + DIR_PATH)
	dbConnect := packages.DB_connect()

	//Запускаем ежедневную очистку в параллельной горутине
	go dailyPurgeWithMutex(dbConnect, deletePeriodDays, DIR_PATH)

	//triggers.AutoPurgeFilesAndDB(dbConnect, deletePeriodDays, DIR_PATH) //Будет создан триггер на каждый день 00:00

	for {
		purgeMutex.Lock()   // === БЛОКИРОВКА: ждём, если сейчас идёт очистка dailyPurgeWithMutex ===
		purgeMutex.Unlock() // сразу отпускаем, но если очистка идёт(dailyPurgeWithMutex) — цикл будет ждать

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

// Функция триггера очистки базы и файлов
func dailyPurgeWithMutex(dbConnect *sql.DB, deletePeriodDays int, dirPath string) {
	for {
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		fmt.Println(nextMidnight)
		time.Sleep(nextMidnight.Sub(now))

		purgeMutex.Lock() //О
		fmt.Println("🕛 Запускаю очистку...")
		triggers.AutoPurgeFilesAndDB(dbConnect, deletePeriodDays, dirPath)
		fmt.Println("✅ Очистка завершена.")
		purgeMutex.Unlock()
	}
}
