package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestConfig_Load(t *testing.T) {

	tests := []struct {
		name    string
		args    []string
		envVars Config
		want    Config
	}{
		{
			name:    "Defaults",
			args:    nil,
			envVars: Config{ServerAddress: "", BaseURLAddress: "", FileStoragePath: "", DatabaseDSN: ""},
			want: Config{
				ServerAddress:   defaultAddress,
				BaseURLAddress:  defaultAddress,
				FileStoragePath: defaultStoragePath,
				DatabaseDSN:     defaultDatabaseDSN(),
			},
		},
		{
			name:    "Override with flags",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt", "-d", "host=local user=u password=p dbname=db"},
			envVars: Config{ServerAddress: "", BaseURLAddress: "", FileStoragePath: ""},
			want: Config{
				ServerAddress:   "localhost:8888",
				BaseURLAddress:  "localhost:8000",
				FileStoragePath: "st/base.txt",
				DatabaseDSN:     "host=local user=u password=p dbname=db",
			},
		},
		{
			name:    "Override with envs",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt", "-d", "host=local user=u password=p dbname=db"},
			envVars: Config{ServerAddress: "localhost:8800", BaseURLAddress: "localhost:8808", FileStoragePath: "./ddd/ttt.txt", DatabaseDSN: "dsn"},
			want:    Config{ServerAddress: "localhost:8800", BaseURLAddress: "localhost:8808", FileStoragePath: "./ddd/ttt.txt", DatabaseDSN: "dsn"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.envVars.ServerAddress)
			t.Setenv("BASE_URL", tt.envVars.BaseURLAddress)
			t.Setenv("FILE_STORAGE_PATH", tt.envVars.FileStoragePath)
			t.Setenv("DATABASE_DSN", tt.envVars.DatabaseDSN)

			c, err := NewConfig("Test", tt.args)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, *c)
		})
	}
}
