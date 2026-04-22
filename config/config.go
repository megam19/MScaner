// Package config — загрузка конфигурации из переменных окружения.
//
// Почему не YAML/JSON-файл? Приложение крутится в Docker, а через
// docker-compose удобно задавать окружение прямо в yaml. Для чего-то
// более сложного (секреты, валидация) стоит присмотреться к spf13/viper
// или caarlos0/env, но пока задача мелкая — справляемся стандартной библиотекой.
//
// Правило: дефолты не должны ломать запуск на свежей системе. Любую опцию
// можно переопределить env-переменной, не меняя код.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config — полный набор настроек приложения. Неизменяемая структура:
// загрузили один раз в main — дальше только читаем.
type Config struct {
	DirPath         string        // папка для сканирования
	FileExtension   string        // расширение отслеживаемых файлов, с точкой
	DatabasePath    string        // путь к файлу SQLite (директория создаётся при старте)
	ScanInterval    time.Duration // период повторного сканирования
	PurgeInterval   time.Duration // период запуска очистки старых записей
	RetentionDays   int           // сколько дней жить «неподвижной» записи до очистки
	HTTPAddr        string        // адрес веб-интерфейса, напр. ":8080"
	HighlightWindow time.Duration // в течение какого окна файл считается «новым» для UI
}

// Load читает текущее окружение и возвращает Config.
// Значения по умолчанию подобраны под типичный docker-compose запуск.
func Load() Config {
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

// getEnv — строковое значение с дефолтом, пустая строка считается «не задано».
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// getEnvInt — целочисленное значение с дефолтом. Некорректное значение
// молча откатывается к fallback (в проде лучше было бы логировать, но
// для этого проекта достаточно).
func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
