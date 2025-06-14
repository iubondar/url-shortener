// Package config предоставляет функциональность для загрузки и управления конфигурацией приложения.
// Поддерживает загрузку конфигурации из переменных окружения и флагов командной строки.
// Приоритет: переменные окружения > флаги командной строки > значения по умолчанию.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/caarlos0/env"
)

// Config представляет структуру конфигурации приложения.
// Все поля могут быть установлены через переменные окружения или флаги командной строки.
type Config struct {
	ServerAddress   string `json:"server_address" env:"SERVER_ADDRESS"`       // адрес, на котором будет запущен сервер
	BaseURLAddress  string `json:"base_url" env:"BASE_URL"`                   // базовый URL для формирования коротких ссылок
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH"` // путь к файлу хранилища
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`           // строка подключения к базе данных
	EnableHTTPS     bool   `json:"enable_https" env:"ENABLE_HTTPS"`           // флаг для включения HTTPS
}

const (
	defaultAddress     = "localhost:8080"
	defaultStoragePath = "./storage/storage.txt"
	localDatabaseDSN   = "host=localhost user=newuser password=password dbname=url_shortener sslmode=disable" // для локальной разработки
)

// NewConfig создает новую конфигурацию приложения.
// Загружает значения из флагов командной строки и переменных окружения.
// Приоритет: переменные окружения > флаги командной строки > значения по умолчанию.
// Возвращает указатель на Config и ошибку, если она возникла.
func NewConfig(progname string, args []string) (Config, error) {
	// Создаем FlagSet и регистрируем все флаги
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	// Создаем временный конфиг для хранения значений из флагов
	var flagValues Config
	var shortConfig, longConfig string

	// Регистрируем все флаги
	flags.StringVar(&flagValues.ServerAddress, "a", "", "address to run server")
	flags.StringVar(&flagValues.BaseURLAddress, "b", "", "base address to construct short URL")
	flags.StringVar(&flagValues.FileStoragePath, "f", "", "path to storage file")
	flags.StringVar(&flagValues.DatabaseDSN, "d", "", "database DSN")
	flags.BoolVar(&flagValues.EnableHTTPS, "s", false, "enable HTTPS")
	flags.StringVar(&shortConfig, "c", "", "config path (short)")
	flags.StringVar(&longConfig, "config", "", "config path (long)")

	// Парсим флаги
	err := flags.Parse(args)
	if err != nil {
		return Config{}, err
	}

	// Получаем путь к конфигурационному файлу
	configPath, err := getConfigPath(shortConfig, longConfig)
	if err != nil {
		return Config{}, err
	}

	// Создаем конфиг из дефолтных значений
	c := Config{
		ServerAddress:   defaultAddress,
		BaseURLAddress:  defaultAddress,
		FileStoragePath: defaultStoragePath,
		DatabaseDSN:     defaultDatabaseDSN(),
		EnableHTTPS:     false,
	}
	if configPath != "" {
		// Пытаемся загрузить из файла
		fc, err := loadConfigFromFile(configPath)
		if err != nil {
			return Config{}, err
		}
		// Перезаписываем значения
		c.overrideWith(fc, true)
	}

	// Перезаписываем значения из флагов, если они заданы
	updateEnableHTTPS := flagValues.EnableHTTPS
	c.overrideWith(flagValues, updateEnableHTTPS)

	// Перезаписываем значения из переменных окружения
	var envValues Config
	err = env.Parse(&envValues)
	if err != nil {
		return Config{}, err
	}
	if _, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		c.ServerAddress = envValues.ServerAddress
	}
	if _, ok := os.LookupEnv("BASE_URL"); ok {
		c.BaseURLAddress = envValues.BaseURLAddress
	}
	if _, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.FileStoragePath = envValues.FileStoragePath
	}
	if _, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.DatabaseDSN = envValues.DatabaseDSN
	}
	if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		c.EnableHTTPS = envValues.EnableHTTPS
	}

	return c, nil
}

// getConfigPath определяет путь к конфигурационному файлу из флагов и переменных окружения.
// Возвращает путь к файлу и ошибку, если заданы оба флага одновременно.
func getConfigPath(shortConfig, longConfig string) (string, error) {
	// Проверяем, что не заданы оба флага одновременно
	if shortConfig != "" && longConfig != "" {
		return "", fmt.Errorf("cannot use both -c and -config flags")
	}

	// Определяем путь к конфигурационному файлу
	if shortConfig != "" {
		return shortConfig, nil
	}
	if longConfig != "" {
		return longConfig, nil
	}
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig, nil
	}

	return "", nil
}

// loadConfigFromFile загружает конфигурацию из файла.
// Возвращает Config и ошибку, если она возникла.
func loadConfigFromFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return Config{}, err
	}

	var c Config
	err = json.Unmarshal(fileBytes, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}

// overrideWith перезаписывает значения конфига значениями переданного конфига если они не пустые
func (c *Config) overrideWith(o Config, updateEnableHTTPS bool) {
	if o.ServerAddress != "" {
		c.ServerAddress = o.ServerAddress
	}
	if o.BaseURLAddress != "" {
		c.BaseURLAddress = o.BaseURLAddress
	}
	if o.FileStoragePath != "" {
		c.FileStoragePath = o.FileStoragePath
	}
	if o.DatabaseDSN != "" {
		c.DatabaseDSN = o.DatabaseDSN
	}
	// Обновляем EnableHTTPS только если updateEnableHTTPS == true
	if updateEnableHTTPS {
		c.EnableHTTPS = o.EnableHTTPS
	}
}

// defaultDatabaseDSN возвращает строку подключения к базе данных по умолчанию.
// Для локальной разработки возвращает предустановленное значение,
// иначе возвращает пустую строку.
func defaultDatabaseDSN() string {
	if isRunningLocally() {
		return localDatabaseDSN
	}

	return ""
}

// isRunningInDocker проверяет, запущено ли приложение в контейнере Docker.
// Возвращает true, если приложение запущено в Docker.
func isRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

// isRunningLocally проверяет, запущено ли приложение локально.
// Возвращает true, если приложение запущено не в Docker.
func isRunningLocally() bool {
	return !isRunningInDocker()
}
