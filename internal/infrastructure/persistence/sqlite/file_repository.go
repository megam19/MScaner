package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"MScaner/internal/domain/file"
)

// FileRepository — реализация file.Repository поверх SQLite.
// Все методы используют ExecContext/QueryContext, чтобы корректно
// прерываться при отмене вызывающего контекста.
type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

// List отдаёт все записи, включая soft-deleted.
// Фильтровать удалённые — это ответственность UI/use-case, а не хранилища.
func (r *FileRepository) List(ctx context.Context) ([]file.File, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT fileName, fileSize, createdAt, updatedAt, deletedAt FROM files`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []file.File
	for rows.Next() {
		var f file.File
		// deletedAt — NULLable, поэтому читаем через sql.NullTime и только
		// при .Valid кладём указатель в доменную модель.
		var deletedAt sql.NullTime
		if err := rows.Scan(&f.Name, &f.Size, &f.CreatedAt, &f.UpdatedAt, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			f.DeletedAt = &t
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// Save — upsert по уникальному полю fileName.
// Важный побочный эффект: при повторной вставке deletedAt сбрасывается в NULL.
// Это поведение «файл воскрес» — он снова появился в папке после пропажи.
func (r *FileRepository) Save(ctx context.Context, f file.File) error {
	const q = `INSERT INTO files (fileName, fileSize) VALUES (?, ?)
		ON CONFLICT(fileName) DO UPDATE SET
			fileSize  = excluded.fileSize,
			updatedAt = CURRENT_TIMESTAMP,
			deletedAt = NULL`
	_, err := r.db.ExecContext(ctx, q, f.Name, f.Size)
	return err
}

// MarkDeleted проставляет deletedAt = CURRENT_TIMESTAMP для переданных имён.
//
// Условие "deletedAt IS NULL" в WHERE защищает время первой пометки:
// если файл уже помечен удалённым, повторный вызов не «обновит» время.
// Это важно, потому что UI/purge ориентируются именно на момент первого удаления.
func (r *FileRepository) MarkDeleted(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}
	// Собираем плейсхолдеры "?,?,?,…" в количестве len(names).
	// Не используем sql.Named/IN(?) напрямую, т.к. database/sql не раскрывает
	// срезы в IN-список автоматически.
	placeholders := strings.Repeat("?,", len(names))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(names))
	for i, n := range names {
		args[i] = n
	}
	q := `UPDATE files SET deletedAt = CURRENT_TIMESTAMP WHERE deletedAt IS NULL AND fileName IN (` + placeholders + `)`
	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}

// DeleteOlderThan физически удаляет записи.
//
// COALESCE(deletedAt, updatedAt) — хитрый трюк: если файл помечен удалённым,
// смотрим на deletedAt; если нет — на updatedAt. Так один запрос покрывает
// оба сценария:
//   - soft-deleted запись простояла дольше retention → чистим;
//   - активная запись давно не менялась → чистим.
//
// Формат "2006-01-02T15:04:05Z" — способ Go хранить ISO-8601; важно именно
// так, потому что колонки в SQLite — TEXT, и сравнение строковое.
func (r *FileRepository) DeleteOlderThan(ctx context.Context, t time.Time) (int64, error) {
	const q = `DELETE FROM files WHERE COALESCE(deletedAt, updatedAt) <= ?`
	res, err := r.db.ExecContext(ctx, q, t.Format("2006-01-02T15:04:05Z"))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
