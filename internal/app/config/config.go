package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURLAddress  string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

const (
	defaultAddress     = "localhost:8080"
	defaultStoragePath = "./storage/storage.txt"
)

var Default Config

func (c *Config) Load(progname string, args []string) (err error) {
	// https://eli.thegreenplace.net/2020/testing-flag-parsing-in-go-programs/
	// Загружаем значения из переданных аргументов командной строки
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	flags.StringVar(&c.ServerAddress, "a", defaultAddress, "address to run server")
	flags.StringVar(&c.BaseURLAddress, "b", defaultAddress, "base address to construct short URL")
	flags.StringVar(&c.FileStoragePath, "f", defaultStoragePath, "path to storage file")

	err = flags.Parse(args)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Переписываем значения из переменных окружения
	err = env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}

	return err
}
