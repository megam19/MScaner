package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"MScaner/internal/domain/file"
)

// PurgeService — use-case «убрать из БД слишком старые записи».
//
// Физически удаляет записи, которые дольше retentionDays не обновлялись
// (активные, но «забытые») и которые были помечены удалёнными дольше этого
// же срока. Сам порог даты считается здесь, но правило отсева «по полю»
// живёт в реализации Repository.DeleteOlderThan (infrastructure).
type PurgeService struct {
	repo          file.Repository
	retentionDays int
}

func NewPurgeService(repo file.Repository, retentionDays int) *PurgeService {
	return &PurgeService{repo: repo, retentionDays: retentionDays}
}

// Run считает границу «старое/новое» относительно текущего момента в UTC
// и отдаёт её репозиторию.
func (s *PurgeService) Run(ctx context.Context) error {
	threshold := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	deleted, err := s.repo.DeleteOlderThan(ctx, threshold)
	if err != nil {
		return fmt.Errorf("purge files: %w", err)
	}
	log.Printf("Purge complete: %d files removed (older than %s)", deleted, threshold.Format(time.RFC3339))
	return nil
}
