package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iubondar/url-shortener/internal/app/storage/testhelpers"
)

// ExampleFileRepository_SaveURL демонстрирует сохранение URL в файловом хранилище.
func ExampleFileRepository_SaveURL() {
	// Создаем временный файл для теста
	tempFile := filepath.Join(os.TempDir(), "example_urls.json")
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			fmt.Printf("Error removing temp file: %v\n", err)
		}
	}()

	// Создаем репозиторий
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

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

// ExampleFileRepository_RetrieveByShortURL демонстрирует получение URL по короткому идентификатору.
func ExampleFileRepository_RetrieveByShortURL() {
	// Создаем временный файл для теста
	tempFile := filepath.Join(os.TempDir(), "example_urls.json")
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			fmt.Printf("Error removing temp file: %v\n", err)
		}
	}()

	// Создаем репозиторий
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

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

// ExampleFileRepository_RetrieveUserURLs демонстрирует получение всех URL пользователя.
func ExampleFileRepository_RetrieveUserURLs() {
	// Создаем временный файл для теста
	tempFile := filepath.Join(os.TempDir(), "example_urls.json")
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			fmt.Printf("Error removing temp file: %v\n", err)
		}
	}()

	// Создаем репозиторий
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Сохраняем несколько URL
	_, _, err = repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.com")
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
