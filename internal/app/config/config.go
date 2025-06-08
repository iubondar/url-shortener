// Package config предоставляет функциональность для загрузки и управления конфигурацией приложения.
// Поддерживает загрузку конфигурации из переменных окружения и флагов командной строки.
// Приоритет: переменные окружения > флаги командной строки > значения по умолчанию.
package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
)

// Config представляет структуру конфигурации приложения.
// Все поля могут быть установлены через переменные окружения или флаги командной строки.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`    // адрес, на котором будет запущен сервер
	BaseURLAddress  string `env:"BASE_URL"`          // базовый URL для формирования коротких ссылок
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // путь к файлу хранилища
	DatabaseDSN     string `env:"DATABASE_DSN"`      // строка подключения к базе данных
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`      // флаг для включения HTTPS
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
	var c Config

	// https://eli.thegreenplace.net/2020/testing-flag-parsing-in-go-programs/
	// Загружаем значения из переданных аргументов командной строки
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	flags.StringVar(&c.ServerAddress, "a", defaultAddress, "address to run server")
	flags.StringVar(&c.BaseURLAddress, "b", defaultAddress, "base address to construct short URL")
	flags.StringVar(&c.FileStoragePath, "f", defaultStoragePath, "path to storage file")
	flags.StringVar(&c.DatabaseDSN, "d", defaultDatabaseDSN(), "database DSN")
	flags.BoolVar(&c.EnableHTTPS, "s", false, "enable HTTPS")

	err := flags.Parse(args)
	if err != nil {
		return Config{}, err
	}

	// Переписываем значения из переменных окружения
	err = env.Parse(&c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
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
