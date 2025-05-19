package config

import (
	"fmt"
	"os"
)

// ExampleNewConfig_withFlags демонстрирует создание конфигурации с использованием флагов командной строки.
func ExampleNewConfig_withFlags() {
	// Создаем конфигурацию с флагами командной строки
	args := []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "custom/path.txt", "-d", "host=local user=u password=p dbname=db"}
	config, err := NewConfig("Example", args)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим значения конфигурации
	fmt.Printf("Server Address: %s\n", config.ServerAddress)
	fmt.Printf("Base URL: %s\n", config.BaseURLAddress)
	fmt.Printf("Storage Path: %s\n", config.FileStoragePath)
	fmt.Printf("Database DSN: %s\n", config.DatabaseDSN)
	// Output:
	// Server Address: localhost:8888
	// Base URL: localhost:8000
	// Storage Path: custom/path.txt
	// Database DSN: host=local user=u password=p dbname=db
}

// ExampleNewConfig_withEnvVars демонстрирует создание конфигурации с использованием переменных окружения.
func ExampleNewConfig_withEnvVars() {
	// Устанавливаем переменные окружения
	os.Setenv("SERVER_ADDRESS", "localhost:9999")
	os.Setenv("BASE_URL", "localhost:9998")
	os.Setenv("FILE_STORAGE_PATH", "env/path.txt")
	os.Setenv("DATABASE_DSN", "host=env user=env password=env dbname=env")
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
	}()

	// Создаем конфигурацию без флагов командной строки
	config, err := NewConfig("Example", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим значения конфигурации
	fmt.Printf("Server Address: %s\n", config.ServerAddress)
	fmt.Printf("Base URL: %s\n", config.BaseURLAddress)
	fmt.Printf("Storage Path: %s\n", config.FileStoragePath)
	fmt.Printf("Database DSN: %s\n", config.DatabaseDSN)
	// Output:
	// Server Address: localhost:9999
	// Base URL: localhost:9998
	// Storage Path: env/path.txt
	// Database DSN: host=env user=env password=env dbname=env
}
