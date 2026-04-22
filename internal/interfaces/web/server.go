// Package web — слой interfaces/presentation: HTTP-сервер, который показывает
// содержимое БД человеку в браузере.
//
// Это «адаптер к внешнему миру»: получает HTTP-запросы, достаёт данные через
// доменный Repository и рендерит HTML-шаблон. Бизнес-правил здесь нет —
// только отображение (и немного логики «как красить строку»).
//
// Шаблон .html встраивается в бинарь через //go:embed, чтобы деплой
// оставался однофайловым — не нужно копировать templates/ рядом с бинарём.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"

	"MScaner/internal/domain/file"
)

// templatesFS содержит все HTML-шаблоны, вшитые в бинарь на этапе компиляции.
// Путь в директиве embed — относительный от расположения этого .go-файла.
//
//go:embed templates/*.html
var templatesFS embed.FS

// Server — маленький HTTP-сервер со своим набором хендлеров.
//
// Конструктор принимает Repository (интерфейс домена), а не конкретный SQLite —
// благодаря этому сервер легко переиспользовать с другой БД или заглушкой в тестах.
type Server struct {
	repo            file.Repository
	tmpl            *template.Template
	addr            string        // адрес прослушивания, напр. ":8080"
	highlightWindow time.Duration // окно, в котором файл считается «новым»
}

// NewServer парсит шаблоны на старте: если HTML сломан — узнаем сразу,
// а не при первом запросе пользователя.
func NewServer(repo file.Repository, addr string, highlightWindow time.Duration) (*Server, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &Server{repo: repo, tmpl: tmpl, addr: addr, highlightWindow: highlightWindow}, nil
}

// Run запускает сервер и блокирует горутину вызова до отмены ctx или ошибки.
//
// При отмене контекста выполняется graceful shutdown: текущие запросы
// получают до 5 секунд, чтобы завершиться, прежде чем сервер принудительно
// закроется.
func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/files", s.handleAPIFiles)

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // защита от slowloris-атак
	}

	// Запускаем ListenAndServe в фоне, чтобы в main-горутине ждать сразу
	// два сигнала: отмену контекста или ошибку сервера.
	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", s.addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// fileRow — view-model для одной строки таблицы. Отличается от доменной
// file.File тем, что содержит уже вычисленный Status — текст для класса
// CSS и для бейджа. Так шаблон остаётся простым (без логики окраски).
type fileRow struct {
	Name      string
	Size      int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Status    string // "new" | "deleted" | ""
}

// buildRows превращает доменные File в view-модели, попутно считает счётчики
// и сортирует: сначала новые, потом обычные, в конце удалённые; внутри групп — по имени.
func (s *Server) buildRows(files []file.File, now time.Time) (rows []fileRow, added, deleted int) {
	rows = make([]fileRow, 0, len(files))
	for _, f := range files {
		row := fileRow{
			Name:      f.Name,
			Size:      f.Size,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
			DeletedAt: f.DeletedAt,
		}
		switch {
		case f.IsDeleted():
			row.Status = "deleted"
			deleted++
		case f.IsNew(now, s.highlightWindow):
			row.Status = "new"
			added++
		}
		rows = append(rows, row)
	}

	statusWeight := func(st string) int {
		switch st {
		case "new":
			return 0
		case "":
			return 1
		default: // "deleted"
			return 2
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		wi, wj := statusWeight(rows[i].Status), statusWeight(rows[j].Status)
		if wi != wj {
			return wi < wj
		}
		return rows[i].Name < rows[j].Name
	})
	return rows, added, deleted
}

// handleIndex — главная страница со списком файлов.
// Принимает только "/", остальное отдаёт 404 — это защищает от коллизий
// с путями, которые кто-то добавит позже ("/api/foo" → должен идти по явному хендлеру).
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files, err := s.repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	rows, added, deleted := s.buildRows(files, now)

	data := struct {
		Rows        []fileRow
		Count       int
		Added       int
		Deleted     int
		Now         time.Time
		WindowHuman string
	}{
		Rows:        rows,
		Count:       len(rows),
		Added:       added,
		Deleted:     deleted,
		Now:         now,
		WindowHuman: s.highlightWindow.String(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		// Здесь ответ уже частично мог уйти клиенту (шаблон пишет в w),
		// поэтому корректного способа отдать 500 нет — только логируем.
		log.Printf("render template: %v", err)
	}
}

// handleAPIFiles — JSON-представление того же списка. Удобно для интеграций
// и для отладки через curl. Структура — прямой JSON-сериализованный file.File.
func (s *Server) handleAPIFiles(w http.ResponseWriter, r *http.Request) {
	files, err := s.repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if files == nil {
		// Возвращаем пустой массив, а не null — так удобнее клиентам.
		files = []file.File{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(files)
}
