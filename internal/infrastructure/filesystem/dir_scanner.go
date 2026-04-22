// Package filesystem — реализация доменного интерфейса file.Scanner поверх
// стандартной файловой системы (os.ReadDir).
//
// Это infrastructure-слой: здесь «грязь реального мира» — чтение диска,
// фильтрация по расширению, работа с единицами (КБ). Домен про это не знает.
// Если появится второй источник (S3, сетевая шара через свой API) — сделайте
// ещё один пакет в infrastructure, реализующий тот же file.Scanner.
package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"MScaner/internal/domain/file"
)

// DirScanner сканирует одну плоскую папку, не заходя в подпапки.
// Фильтрует по одному расширению (например ".mxf").
type DirScanner struct {
	extension string // с точкой, в нижнем регистре
}

// NewDirScanner. Расширение приводится к нижнему регистру, чтобы сравнение
// было регистронезависимым (".MXF" === ".mxf").
func NewDirScanner(extension string) *DirScanner {
	return &DirScanner{extension: strings.ToLower(extension)}
}

// Scan читает содержимое директории и возвращает список подходящих файлов.
//
// Ошибки на отдельных файлах (не удалось получить stat) не валят весь скан —
// такой файл просто пропускается. Ошибка возвращается только если
// не удалось прочитать саму директорию.
//
// Размер переводится в килобайты (/ 1024). Это исторически установленный
// формат; при необходимости легко поменять здесь, не трогая домен.
func (s *DirScanner) Scan(ctx context.Context, path string) ([]file.File, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", path, err)
	}

	var result []file.File
	for _, entry := range entries {
		// Уважаем контекст: если вызывающая сторона отменила операцию
		// (таймаут, SIGTERM), прерываем цикл.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != s.extension {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			// Файл мог исчезнуть между ReadDir и Info — это нормально
			// для живой папки, просто пропускаем.
			continue
		}
		result = append(result, file.File{
			Name: entry.Name(),
			Size: info.Size() / 1024,
		})
	}
	return result, nil
}
