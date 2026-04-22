package file

// DiffResult — результат сравнения состояния БД с текущим состоянием папки.
//
// Разделение на Upserts и Disappeared позволяет приложению применить
// изменения двумя разными операциями хранилища (Save + MarkDeleted),
// не смешивая их.
type DiffResult struct {
	// Upserts — файлы, которые нужно сохранить:
	//   - впервые увиденные,
	//   - с изменившимся размером,
	//   - ранее помеченные удалёнными, но вновь появившиеся в папке.
	Upserts []File

	// Disappeared — имена файлов, которые присутствовали в БД как активные,
	// но сейчас отсутствуют в папке. Их нужно пометить удалёнными.
	Disappeared []string
}

// Diff — чистая функция: нет побочных эффектов, не ходит в БД, не читает диск.
// Такие функции удобно тестировать: на вход срезы, на выход структура.
//
// existing — текущее состояние БД (включая soft-deleted),
// current  — то, что сейчас реально лежит в папке.
//
// Правило: если файл уже был в БД, помечен как удалённый, и вдруг снова
// появился — он попадает в Upserts, чтобы слой persistence снял метку.
func Diff(existing, current []File) DiffResult {
	// Индексируем для O(1) поиска; без этого на больших папках
	// (десятки тысяч файлов) получили бы O(n*m).
	existingByName := make(map[string]File, len(existing))
	for _, f := range existing {
		existingByName[f.Name] = f
	}
	currentByName := make(map[string]struct{}, len(current))
	for _, f := range current {
		currentByName[f.Name] = struct{}{}
	}

	var res DiffResult

	// Проход 1: что в папке — сверяем с БД, ищем новые/изменённые/воскресшие.
	for _, f := range current {
		prev, ok := existingByName[f.Name]
		if !ok || prev.Size != f.Size || prev.IsDeleted() {
			res.Upserts = append(res.Upserts, f)
		}
	}

	// Проход 2: что в БД — сверяем с папкой, ищем пропавшие.
	for _, f := range existing {
		if f.IsDeleted() {
			continue // уже помечены удалёнными, повторно трогать не нужно
		}
		if _, ok := currentByName[f.Name]; !ok {
			res.Disappeared = append(res.Disappeared, f.Name)
		}
	}
	return res
}
