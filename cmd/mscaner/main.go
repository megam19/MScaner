// Package main — точка входа и composition root проекта MScaner.
//
// Это ЕДИНСТВЕННОЕ место, где соединяются конкретные реализации из
// infrastructure и interfaces с абстракциями из domain. Если вам нужно
// подменить БД, сканер или добавить новый интерфейс (например, gRPC) —
// правка должна быть здесь, а не в глубине сервисов.
//
// Схема зависимостей:
//
//	domain  ←—  application  ←—  infrastructure
//	   ↑                                  ↑
//	   └——————— interfaces/web ———————————┘
//	                   ↑
//	                 main (собирает всё)
//
// Цикл работы приложения:
//
//  1. загрузить конфиг;
//  2. открыть БД (sqlite);
//  3. собрать репозиторий и сканер;
//  4. собрать use-case'ы (ScanService, PurgeService);
//  5. запустить HTTP-сервер и фоновые тикеры scan/purge;
//  6. по SIGINT/SIGTERM — graceful shutdown.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"MScaner/config"
	"MScaner/internal/application"
	"MScaner/internal/infrastructure/filesystem"
	"MScaner/internal/infrastructure/persistence/sqlite"
	"MScaner/internal/interfaces/web"
)

func main() {
	cfg := config.Load()
	log.Printf("MScaner starting: dir=%s ext=%s interval=%s http=%s",
		cfg.DirPath, cfg.FileExtension, cfg.ScanInterval, cfg.HTTPAddr)

	// 1. Инфраструктура: соединение с БД. Держим до конца main.
	db, err := sqlite.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	// 2. Реализации доменных интерфейсов.
	repo := sqlite.NewFileRepository(db)
	scanner := filesystem.NewDirScanner(cfg.FileExtension)

	// 3. Use-case'ы. Они знают только про интерфейсы, не про sqlite/filesystem.
	scanSvc := application.NewScanService(scanner, repo, cfg.DirPath)
	purgeSvc := application.NewPurgeService(repo, cfg.RetentionDays)

	// 4. Веб-интерфейс. Ему нужен Repository для чтения и окно «новизны» для подсветки.
	server, err := web.NewServer(repo, cfg.HTTPAddr, cfg.HighlightWindow)
	if err != nil {
		log.Fatalf("init web server: %v", err)
	}

	// 5. Общий контекст: отменяется при SIGINT/SIGTERM — так docker stop и Ctrl+C
	// корректно гасят и веб-сервер, и фоновые тикеры.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go runPeriodic(ctx, cfg.ScanInterval, "scan", scanSvc.Run)
	go runPeriodic(ctx, cfg.PurgeInterval, "purge", purgeSvc.Run)

	// Веб-сервер запускаем в главной горутине: он заблокирует main до отмены ctx.
	if err := server.Run(ctx); err != nil {
		log.Printf("http server: %v", err)
	}
	log.Println("MScaner stopped")
}

// runPeriodic выполняет fn сразу, затем повторяет по таймеру, пока ctx не отменён.
//
// Почему не time.Tick? Потому что Ticker можно остановить и освободить ресурсы;
// Tick (в отличие от NewTicker) остановить нельзя — это типичная утечка в Go.
//
// Почему сначала один запуск, потом тикер? Чтобы не ждать первый интервал
// на старте — пользователю приятно видеть данные сразу.
func runPeriodic(ctx context.Context, interval time.Duration, name string, fn func(context.Context) error) {
	run := func() {
		if err := fn(ctx); err != nil {
			log.Printf("%s failed: %v", name, err)
		}
	}

	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
