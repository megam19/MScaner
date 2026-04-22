// Package config — загрузка конфигурации из переменных окружения и/или
// из файла config.env рядом с исполняемым файлом.
//
// Приоритет источников (сверху вниз):
//
//  1. Реальная переменная окружения процесса.
//  2. config.env в той же папке, где лежит mscaner(.exe).
//  3. config.env в текущей рабочей директории.
//  4. Дефолт, прошитый в код.
//
// Такой порядок удобен и для Docker (задаём env в compose), и для
// Windows-деплоя (рядом с .exe кладётся config.env и редактируется в
// блокноте).
package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config — полный набор настроек приложения. Неизменяемая структура:
// загрузили один раз в main — дальше только читаем.
type Config struct {
	DirPath         string        // папка для сканирования (на Windows — UNC, например \\server\share\folder)
	FileExtension   string        // расширение отслеживаемых файлов, с точкой
	DatabasePath    string        // путь к файлу SQLite (директория создаётся при старте)
	ScanInterval    time.Duration // период повторного сканирования
	PurgeInterval   time.Duration // период запуска очистки старых записей
	RetentionDays   int           // сколько дней жить «неподвижной» записи до очистки
	HTTPAddr        string        // адрес веб-интерфейса, напр. ":8080"
	HighlightWindow time.Duration // в течение какого окна файл считается «новым» для UI
}

// Load читает текущее окружение и возвращает Config.
func Load() Config {
	// Сначала пробуем файл рядом с бинарём — это основной способ для
	// Windows-деплоя, когда пользователь просто положил mscaner.exe и
	// отредактировал config.env по соседству.
	if exe, err := os.Executable(); err == nil {
		loadEnvFile(filepath.Join(filepath.Dir(exe), "config.env"))
	}
	// Затем — файл в текущей рабочей директории (удобно при dev-запуске).
	loadEnvFile("config.env")

	return Config{
		DirPath:         getEnv("DIR_PATH", "/scan"),
		FileExtension:   getEnv("FILE_EXTENSION", ".mxf"),
		DatabasePath:    getEnv("DATABASE_PATH", "./database/sqlite3DB.db"),
		ScanInterval:    time.Duration(getEnvInt("SCAN_INTERVAL_SEC", 120)) * time.Second,
		PurgeInterval:   time.Duration(getEnvInt("PURGE_INTERVAL_HOURS", 24)) * time.Hour,
		RetentionDays:   getEnvInt("RETENTION_DAYS", 10),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		HighlightWindow: time.Duration(getEnvInt("HIGHLIGHT_WINDOW_SEC", 3600)) * time.Second,
	}
}

// loadEnvFile пробует прочитать .env-подобный файл и выставить из него
// переменные окружения, но ТОЛЬКО те, которые ещё не заданы реальным
// окружением процесса. Таким образом «настоящая» env всегда побеждает.
//
// Формат:
//
//   - KEY=VALUE                            — обычная пара
//   - # комментарий                        — игнорируется
//   - пустые строки                        — игнорируются
//   - KEY="значение с пробелами"           — двойные кавычки снимаются
//
// Значения НЕ экранируются: обратный слэш остаётся обратным слэшом,
// поэтому UNC-пути вида \\server\share записываются как есть.
// Ошибка открытия файла намеренно проглатывается — файл может отсутствовать.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Снимаем обрамляющие двойные кавычки, если они есть.
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		// Реальный env имеет приоритет: если переменная уже задана —
		// файл её НЕ перекрывает.
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

// getEnv — строковое значение с дефолтом, пустая строка считается «не задано».
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// getEnvInt — целочисленное значение с дефолтом. Некорректное значение
// молча откатывается к fallback.
func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
