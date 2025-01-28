package config

import "flag"

type Config struct {
	ServerAddress  string
	BaseURLAddress string
}

var Default Config

func init() {
	flag.StringVar(&Default.ServerAddress, "a", "localhost:8080", "address to run server")
	flag.StringVar(&Default.BaseURLAddress, "b", "localhost:8080", "base address to construct short URL")
}
