package packages

// Функция возвращает элементы, которые:
// - есть в folder, но нет в db   → новые
// - есть в обоих, но размер отличается → изменённые
func DifferentsToWriteDB(db []ItemStruct, folder []ItemStruct) []ItemStruct {
	// Карта для быстрого поиска: имя файла → размер в базе
	dbMap := make(map[string]int64, len(db))

	for _, item := range db {
		dbMap[item.FileName] = item.FileSize
	}

	var results []ItemStruct

	// Проходим по текущему состоянию папки
	for _, file := range folder {
		if dbSize, exists := dbMap[file.FileName]; !exists || dbSize != file.FileSize {
			// либо файла вообще нет в базе
			// либо размер изменился
			results = append(results, file)
		}
	}

	return results
}
