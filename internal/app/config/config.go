package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURLAddress  string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

const (
	defaultAddress     = "localhost:8080"
	defaultStoragePath = "./storage/storage.txt"
	localDatabaseDSN   = "host=localhost user=newuser password=password dbname=url_shortener sslmode=disable" // для локальной разработки
)

func NewConfig(progname string, args []string) (*Config, error) {
	var c Config

	// https://eli.thegreenplace.net/2020/testing-flag-parsing-in-go-programs/
	// Загружаем значения из переданных аргументов командной строки
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	flags.StringVar(&c.ServerAddress, "a", defaultAddress, "address to run server")
	flags.StringVar(&c.BaseURLAddress, "b", defaultAddress, "base address to construct short URL")
	flags.StringVar(&c.FileStoragePath, "f", defaultStoragePath, "path to storage file")
	flags.StringVar(&c.DatabaseDSN, "d", defaultDatabaseDSN(), "database DSN")

	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}

	// Переписываем значения из переменных окружения
	err = env.Parse(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func defaultDatabaseDSN() string {
	if isRunningLocally() {
		return localDatabaseDSN
	}

	return ""
}

func isRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func isRunningLocally() bool {
	return !isRunningInDocker()
}
