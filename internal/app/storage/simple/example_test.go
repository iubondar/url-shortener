package simple

import (
	"context"
	"fmt"

	"github.com/iubondar/url-shortener/internal/app/storage/testhelpers"
)

// ExampleSimpleRepository_SaveURL демонстрирует сохранение URL в хранилище.
func ExampleSimpleRepository_SaveURL() {
	// Создаем новое хранилище
	repo := NewSimpleRepository()

	// Сохраняем URL
	id, exists, err := repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим результат
	fmt.Printf("ID length: %d, Exists: %v\n", len(id), exists)
	// Output: ID length: 8, Exists: false
}

// ExampleSimpleRepository_RetrieveByShortURL демонстрирует получение URL по короткому идентификатору.
func ExampleSimpleRepository_RetrieveByShortURL() {
	// Создаем новое хранилище
	repo := NewSimpleRepository()

	// Сохраняем URL
	id, _, err := repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Получаем запись по короткому идентификатору
	record, err := repo.RetrieveByShortURL(context.Background(), id)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим результат
	fmt.Printf("Original URL: %s\n", record.OriginalURL)
	// Output: Original URL: http://example.com
}

// ExampleSimpleRepository_RetrieveUserURLs демонстрирует получение всех URL пользователя.
func ExampleSimpleRepository_RetrieveUserURLs() {
	// Создаем новое хранилище
	repo := NewSimpleRepository()

	// Сохраняем несколько URL
	_, _, err := repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_, _, err = repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.org")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Получаем все URL пользователя
	records, err := repo.RetrieveUserURLs(context.Background(), testhelpers.TestUUID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим количество найденных URL
	fmt.Printf("Found %d URLs\n", len(records))
	// Output: Found 2 URLs
}
