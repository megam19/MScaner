// Package sqlite — реализация persistence-слоя поверх SQLite.
//
// Используется драйвер modernc.org/sqlite — это чистый Go, без CGO,
// поэтому сборка кросс-платформенна и не требует gcc внутри контейнера.
//
// Этот пакет — единственное место во всём проекте, где допустимо писать SQL.
// Если SQL просочился выше (в application или interfaces) — это ошибка:
// переносите в реализацию Repository.
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // регистрирует драйвер "sqlite" в database/sql
)

// schema — начальная схема. Применяется на каждый старт через CREATE TABLE
// IF NOT EXISTS, так что для чистой БД это и есть вся инициализация.
const schema = `CREATE TABLE IF NOT EXISTS files (
	fileName  TEXT NOT NULL UNIQUE,
	fileSize  INTEGER,
	createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
	updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
	deletedAt DATETIME
);`

// Open открывает (или создаёт) БД, создаёт директорию для файла, применяет
// схему и запускает миграции для старых БД.
func Open(path string) (*sql.DB, error) {
	// Создаём директорию для файла БД, если её ещё нет. Без этого SQLite
	// откажется создавать файл в несуществующей папке.
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// migrate добавляет колонки, появившиеся в новых версиях, для уже
// существующих БД. SQLite НЕ поддерживает "ADD COLUMN IF NOT EXISTS",
// поэтому ошибку «duplicate column» мы сознательно проглатываем —
// это значит, что миграция уже была применена.
//
// При добавлении новой миграции в будущем:
//  1. обновите SQL в schema (для свежих БД);
//  2. добавьте ALTER TABLE сюда (для старых БД);
//  3. обработайте ожидаемую ошибку «уже существует».
func migrate(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE files ADD COLUMN deletedAt DATETIME`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}
