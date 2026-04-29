package packages

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func ScanDir(dirPath string) []ItemStruct {

	var folderInfo_struct ItemStruct
	var Arr_folderInfo_structs []ItemStruct

	files, err := os.ReadDir(dirPath) // Читаем все файлы по пути dirPath(Расчет на 20-40к файлов)
	if err != nil {
		fmt.Println("Ошибка при чтении директории: ", dirPath, err)
	}

	//Читаем название,размер файлов, только .mxf формат, остальное отбрасываем
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

		folderInfo_struct = ItemStruct{file.Name(), size}                          //наполняем структуру
		Arr_folderInfo_structs = append(Arr_folderInfo_structs, folderInfo_struct) // наполняем массив структурами
	}
	log.Println("Завершена сканирование папки:", "Forward2")
	return Arr_folderInfo_structs
}
