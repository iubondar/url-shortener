package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress  string
	BaseURLAddress string
}

const defaultAddress = "localhost:8080"

var Default Config

func init() {
	flag.StringVar(&Default.ServerAddress, "a", defaultAddress, "address to run server")
	flag.StringVar(&Default.BaseURLAddress, "b", defaultAddress, "base address to construct short URL")
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
