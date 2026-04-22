package file

import (
	"context"
	"time"
)

// Repository — доменный интерфейс хранилища файлов.
//
// Ключевая идея DDD: домен ОБЪЯВЛЯЕТ, что ему нужно от хранилища,
// но НЕ знает, как оно устроено (SQLite, Postgres, память — не важно).
// Реализация живёт в internal/infrastructure/persistence/... и подключается
// в composition root (cmd/mscaner/main.go).
//
// Если вам нужен новый способ достать или изменить данные — добавьте сюда
// метод, затем реализуйте его в infrastructure. НЕ добавляйте сюда SQL,
// ORM-теги или что-либо специфичное для конкретной СУБД.
type Repository interface {
	// List возвращает все записи, включая помеченные удалёнными.
	// UI сам решает, какие показывать и как окрашивать.
	List(ctx context.Context) ([]File, error)

	// Save вставляет или обновляет запись по уникальному имени.
	// Если файл был помечен удалённым — метка снимается (файл «воскрес»).
	Save(ctx context.Context, f File) error

	// MarkDeleted проставляет время удаления для перечисленных имён.
	// Повторный вызов на уже удалённом не должен менять время удаления.
	MarkDeleted(ctx context.Context, names []string) error

	// DeleteOlderThan физически удаляет записи, не изменявшиеся и не
	// помеченные удалёнными дольше порога t. Возвращает кол-во удалённых строк.
	DeleteOlderThan(ctx context.Context, t time.Time) (int64, error)
}
