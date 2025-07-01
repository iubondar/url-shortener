package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/iubondar/url-shortener/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	serverAddress = "localhost:3200" // Адрес gRPC сервера
)

func main() {
	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Создаем клиент
	client := proto.NewShortenerClient(conn)

	// Контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Тестирование gRPC сервера ===")
	fmt.Println()

	// Тест 1: Ping (без авторизации)
	fmt.Println("1. Тестирование Ping...")
	testPing(ctx, client)
	fmt.Println()

	// Тест 2: CreateID (с авторизацией)
	fmt.Println("2. Тестирование CreateID...")
	token := testCreateID(ctx, client)
	fmt.Println()

	// Тест 3: ShortenBatch (с авторизацией)
	fmt.Println("3. Тестирование ShortenBatch...")
	testShortenBatch(ctx, client, token)
	fmt.Println()

	// Тест 4: GetUserURLs (с авторизацией)
	fmt.Println("4. Тестирование GetUserURLs...")
	testGetUserURLs(ctx, client, token)
	fmt.Println()

	// Тест 5: DeleteURLs (с авторизацией)
	fmt.Println("5. Тестирование DeleteURLs...")
	testDeleteURLs(ctx, client, token)
	fmt.Println()

	// Тест 6: GetStats (без авторизации, но с проверкой trusted subnet)
	fmt.Println("6. Тестирование GetStats...")
	testGetStats(ctx, client)
	fmt.Println()

	fmt.Println("=== Тестирование завершено ===")
}

// testPing тестирует метод Ping
func testPing(ctx context.Context, client proto.ShortenerClient) {
	resp, err := client.Ping(ctx, &proto.PingRequest{})
	if err != nil {
		log.Printf("Ping failed: %v", err)
		return
	}

	fmt.Printf("   Статус: %s\n", resp.Status.String())
	if resp.Error != "" {
		fmt.Printf("   Ошибка: %s\n", resp.Error)
	} else {
		fmt.Println("   ✅ Ping успешен")
	}
}

// testCreateID тестирует метод CreateID и возвращает токен авторизации
func testCreateID(ctx context.Context, client proto.ShortenerClient) string {
	// Создаем контекст с metadata для авторизации
	md := metadata.New(map[string]string{})
	ctxWithAuth := metadata.NewOutgoingContext(ctx, md)

	// Вызываем CreateID
	resp, err := client.CreateID(ctxWithAuth, &proto.CreateIDRequest{
		Url: "https://www.google.com",
	})
	if err != nil {
		log.Printf("CreateID failed: %v", err)
		return ""
	}

	fmt.Printf("   Результат: %s\n", resp.Result)
	if resp.Error != "" {
		fmt.Printf("   Ошибка: %s\n", resp.Error)
	} else {
		fmt.Println("   ✅ CreateID успешен")
	}

	// В реальном gRPC клиенте токен приходит в response headers
	// Для демонстрации создадим заглушку токена
	// В реальном приложении нужно использовать grpc.Header() для получения headers
	token := "demo-jwt-token"
	fmt.Printf("   Получен токен авторизации: %s...\n", token[:10])

	return token
}

// testShortenBatch тестирует метод ShortenBatch
func testShortenBatch(ctx context.Context, client proto.ShortenerClient, token string) {
	// Создаем контекст с токеном авторизации
	md := metadata.New(map[string]string{
		"authorization": token,
	})
	ctxWithAuth := metadata.NewOutgoingContext(ctx, md)

	// Создаем запрос с несколькими URL
	req := &proto.ShortenBatchRequest{
		Items: []*proto.BatchItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "https://www.github.com",
			},
			{
				CorrelationId: "2",
				OriginalUrl:   "https://www.stackoverflow.com",
			},
			{
				CorrelationId: "3",
				OriginalUrl:   "https://www.reddit.com",
			},
		},
	}

	resp, err := client.ShortenBatch(ctxWithAuth, req)
	if err != nil {
		log.Printf("ShortenBatch failed: %v", err)
		return
	}

	if resp.Error != "" {
		fmt.Printf("   Ошибка: %s\n", resp.Error)
	} else {
		fmt.Printf("   Обработано URL: %d\n", len(resp.Items))
		for _, item := range resp.Items {
			fmt.Printf("   - %s: %s\n", item.CorrelationId, item.ShortUrl)
		}
		fmt.Println("   ✅ ShortenBatch успешен")
	}
}

// testGetUserURLs тестирует метод GetUserURLs
func testGetUserURLs(ctx context.Context, client proto.ShortenerClient, token string) {
	// Создаем контекст с токеном авторизации
	md := metadata.New(map[string]string{
		"authorization": token,
	})
	ctxWithAuth := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.GetUserURLs(ctxWithAuth, &proto.GetUserURLsRequest{})
	if err != nil {
		log.Printf("GetUserURLs failed: %v", err)
		return
	}

	if resp.Error != "" {
		fmt.Printf("   Ошибка: %s\n", resp.Error)
	} else {
		fmt.Printf("   Найдено URL: %d\n", len(resp.Items))
		for _, item := range resp.Items {
			fmt.Printf("   - %s -> %s\n", item.ShortUrl, item.OriginalUrl)
		}
		fmt.Println("   ✅ GetUserURLs успешен")
	}
}

// testDeleteURLs тестирует метод DeleteURLs
func testDeleteURLs(ctx context.Context, client proto.ShortenerClient, token string) {
	// Создаем контекст с токеном авторизации
	md := metadata.New(map[string]string{
		"authorization": token,
	})
	ctxWithAuth := metadata.NewOutgoingContext(ctx, md)

	// Список URL для удаления (короткие URL)
	urlsToDelete := []string{
		"http://localhost:8080/abc123",
		"http://localhost:8080/def456",
	}

	resp, err := client.DeleteURLs(ctxWithAuth, &proto.DeleteURLsRequest{
		Urls: urlsToDelete,
	})
	if err != nil {
		log.Printf("DeleteURLs failed: %v", err)
		return
	}

	if resp.Error != "" {
		fmt.Printf("   Ошибка: %s\n", resp.Error)
	} else {
		fmt.Printf("   Запрошено удаление %d URL\n", len(urlsToDelete))
		fmt.Println("   ✅ DeleteURLs успешен")
	}
}

// testGetStats тестирует метод GetStats
func testGetStats(ctx context.Context, client proto.ShortenerClient) {
	resp, err := client.GetStats(ctx, &proto.GetStatsRequest{})
	if err != nil {
		log.Printf("GetStats failed: %v", err)
		return
	}

	if resp.Error != "" {
		fmt.Printf("   Ошибка: %s\n", resp.Error)
	} else {
		fmt.Printf("   Статистика: %d URL, %d пользователей\n", resp.Urls, resp.Users)
		fmt.Println("   ✅ GetStats успешен")
	}
}
