package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/iubondar/url-shortener/proto"
)

func main() {
	// Подключаемся к gRPC серверу
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Создаем клиент
	client := proto.NewCreateIDHandlerClient(conn)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Первый запрос без токена (должен создать новый)
	fmt.Println("=== Первый запрос без токена ===")
	resp1, err := client.CreateID(ctx, &proto.CreateIDRequest{
		Url: "https://www.google.com",
	})
	if err != nil {
		log.Fatalf("Failed to call CreateID: %v", err)
	}

	fmt.Printf("Response: %+v\n", resp1)

	// Получаем токен из metadata ответа
	// В реальном клиенте нужно извлечь токен из response metadata
	// Для демонстрации используем фиктивный токен
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiIxMjM0NTY3ODkwLTEyMzQtMTIzNC0xMjM0LTEyMzQ1Njc4OTAwIn0.example"

	// Второй запрос с токеном
	fmt.Println("\n=== Второй запрос с токеном ===")
	md := metadata.New(map[string]string{
		"authorization": token,
	})
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp2, err := client.CreateID(ctxWithToken, &proto.CreateIDRequest{
		Url: "https://www.github.com",
	})
	if err != nil {
		log.Fatalf("Failed to call CreateID: %v", err)
	}

	fmt.Printf("Response: %+v\n", resp2)

	// Третий запрос с невалидным токеном (должен создать новый)
	fmt.Println("\n=== Третий запрос с невалидным токеном ===")
	mdInvalid := metadata.New(map[string]string{
		"authorization": "invalid_token",
	})
	ctxWithInvalidToken := metadata.NewOutgoingContext(ctx, mdInvalid)

	resp3, err := client.CreateID(ctxWithInvalidToken, &proto.CreateIDRequest{
		Url: "https://www.stackoverflow.com",
	})
	if err != nil {
		log.Fatalf("Failed to call CreateID: %v", err)
	}

	fmt.Printf("Response: %+v\n", resp3)
}
