package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// setupTestFile создает временный файл для тестов
func setupTestFile(t *testing.B) string {
	tempFile := filepath.Join(os.TempDir(), "benchmark_urls.json")
	t.Cleanup(func() {
		os.Remove(tempFile)
	})
	return tempFile
}

// BenchmarkFileRepository_SaveURL измеряет производительность сохранения URL
func BenchmarkFileRepository_SaveURL(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}
}

// BenchmarkFileRepository_RetrieveByShortURL измеряет производительность получения URL по короткому идентификатору
func BenchmarkFileRepository_RetrieveByShortURL(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveByShortURL(ctx, id)
	}
}

// BenchmarkFileRepository_RetrieveUserURLs измеряет производительность получения всех URL пользователя
func BenchmarkFileRepository_RetrieveUserURLs(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	for range 100 {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveUserURLs(ctx, userID)
	}
}

// BenchmarkFileRepository_DeleteByShortURLs измеряет производительность удаления URL
func BenchmarkFileRepository_DeleteByShortURLs(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	var shortURLs []string
	for i := 0; i < 100; i++ {
		id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")
		shortURLs = append(shortURLs, id)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.DeleteByShortURLs(ctx, userID, shortURLs)
	}
}

// BenchmarkFileRepository_SaveURLs измеряет производительность пакетного сохранения URL
func BenchmarkFileRepository_SaveURLs(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	urls := make([]string, 100)
	for range 100 {
		urls = append(urls, "http://example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.SaveURLs(ctx, urls)
	}
}

// BenchmarkFileRepository_CheckStatus измеряет производительность проверки состояния хранилища
func BenchmarkFileRepository_CheckStatus(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.CheckStatus(ctx)
	}
}
