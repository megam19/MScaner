// Package application — прикладной слой (use-cases) в DDD.
//
// Здесь описаны сценарии работы системы: «просканировать папку и обновить БД»,
// «удалить устаревшие записи». Сервисы оркеструют домен и инфраструктуру,
// но не содержат бизнес-правил (те живут в domain) и не знают про конкретные
// технологии (SQLite, HTTP — это infrastructure/interfaces).
//
// Типичный use-case:
//  1. получить данные через интерфейсы домена (Repository, Scanner);
//  2. вызвать чистую доменную логику (Diff);
//  3. применить изменения через интерфейсы.
package application

import (
	"context"
	"fmt"

	"MScaner/internal/domain/file"
)

// ScanService — use-case «синхронизировать БД с состоянием папки».
//
// Зависит только от доменных интерфейсов. Конкретные реализации
// (sqlite.FileRepository, filesystem.DirScanner) подставляются в main.go.
// Благодаря этому сервис можно тестировать, подсунув in-memory заглушки.
type ScanService struct {
	scanner file.Scanner    // откуда читаем файлы
	repo    file.Repository // куда сохраняем состояние
	path    string          // что именно сканируем (см. Scanner.Scan)
}

// NewScanService — конструктор. Все зависимости явные (Dependency Injection):
// никакой магии, никаких глобальных переменных.
func NewScanService(scanner file.Scanner, repo file.Repository, path string) *ScanService {
	return &ScanService{scanner: scanner, repo: repo, path: path}
}

// Run выполняет один цикл синхронизации: читает папку, читает БД,
// считает разницу и применяет изменения.
//
// Вызывается снаружи по таймеру (см. main.runPeriodic).
// Возвращаемая ошибка — первая «фатальная» (сломалась БД/сканер);
// типичные ошибки на отдельных файлах глотаются сканером, чтобы
// один битый файл не валил весь цикл.
func (s *ScanService) Run(ctx context.Context) error {
	current, err := s.scanner.Scan(ctx, s.path)
	if err != nil {
		return fmt.Errorf("scan directory: %w", err)
	}

	existing, err := s.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("read repository: %w", err)
	}

	// Вся логика «что изменилось» — в доменной функции Diff.
	// Этот сервис не знает правил сравнения, он только оркеструет.
	diff := file.Diff(existing, current)

	for _, f := range diff.Upserts {
		if err := s.repo.Save(ctx, f); err != nil {
			return fmt.Errorf("save file %s: %w", f.Name, err)
		}
	}
	if err := s.repo.MarkDeleted(ctx, diff.Disappeared); err != nil {
		return fmt.Errorf("mark deleted: %w", err)
	}
	return nil
}
