package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ExampleNewConfig демонстрирует создание конфигурации с использованием значений по умолчанию.
func ExampleNewConfig() {
	// Создаем конфигурацию без дополнительных параметров
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
	// Server Address: localhost:8080
	// Base URL: localhost:8080
	// Storage Path: ./storage/storage.txt
	// Database DSN: host=localhost user=newuser password=password dbname=url_shortener sslmode=disable
}

// ExampleNewConfig_withFlags демонстрирует создание конфигурации с использованием флагов командной строки.
func ExampleNewConfig_withFlags() {
	// Создаем конфигурацию с флагами командной строки
	args := []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "custom/path.txt"}
	config, err := NewConfig("Example", args)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим значения конфигурации
	fmt.Printf("Server Address: %s\n", config.ServerAddress)
	fmt.Printf("Base URL: %s\n", config.BaseURLAddress)
	fmt.Printf("Storage Path: %s\n", config.FileStoragePath)
	// Output:
	// Server Address: localhost:8888
	// Base URL: localhost:8000
	// Storage Path: custom/path.txt
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
