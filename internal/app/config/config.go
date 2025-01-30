package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress  string
	BaseURLAddress string
}

var Default Config

func init() {
	flag.StringVar(&Default.ServerAddress, "a", "localhost:8080", "address to run server")
	flag.StringVar(&Default.BaseURLAddress, "b", "localhost:8080", "base address to construct short URL")
}

func (c *Config) Load() {
	flag.Parse()

	// переопределяем конфигурацию из переменных окружения, если они заданы
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		Default.ServerAddress = envRunAddr
	}

	if envRunAddr := os.Getenv("BASE_URL"); envRunAddr != "" {
		Default.BaseURLAddress = envRunAddr
	}
}
