package packages

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func ScanDir(dirPath string) {

	var i_folder ItemStruct
	Arr_items_FolderStruct = Arr_items_FolderStruct[:0] //обнуление слайса но остается прежняя capacity

	files, err := os.ReadDir(dirPath) // Читаем все файлы по пути dirPath(Расчет на 20-40к файлов)
	if err != nil {
		fmt.Println(err)
	}

	//Читаем название,размер файлов только .mxf формат, остаьное отбрасываем
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

		i_folder = ItemStruct{file.Name(), size}                          //наполняем структуру
		Arr_items_FolderStruct = append(Arr_items_FolderStruct, i_folder) // наполняем массив структурами
	}

}
