package simple

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// BenchmarkSimpleRepository_SaveURL измеряет производительность сохранения URL
func BenchmarkSimpleRepository_SaveURL(b *testing.B) {
	repo := NewSimpleRepository()
	userID := uuid.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}
}

// BenchmarkSimpleRepository_RetrieveByShortURL измеряет производительность получения URL по короткому идентификатору
func BenchmarkSimpleRepository_RetrieveByShortURL(b *testing.B) {
	repo := NewSimpleRepository()
	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveByShortURL(ctx, id)
	}
}

// BenchmarkSimpleRepository_RetrieveUserURLs измеряет производительность получения всех URL пользователя
func BenchmarkSimpleRepository_RetrieveUserURLs(b *testing.B) {
	repo := NewSimpleRepository()
	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	for i := 0; i < 100; i++ {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveUserURLs(ctx, userID)
	}
}

// BenchmarkSimpleRepository_DeleteByShortURLs измеряет производительность удаления URL
func BenchmarkSimpleRepository_DeleteByShortURLs(b *testing.B) {
	repo := NewSimpleRepository()
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

// BenchmarkSimpleRepository_SaveURLs измеряет производительность пакетного сохранения URL
func BenchmarkSimpleRepository_SaveURLs(b *testing.B) {
	repo := NewSimpleRepository()
	ctx := context.Background()

	urls := make([]string, 100)
	for i := 0; i < 100; i++ {
		urls[i] = "http://example.com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.SaveURLs(ctx, urls)
	}
}
