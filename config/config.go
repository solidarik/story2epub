package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	DbPath        string   `env:"DB_PATH,required"`
	AllowedDomain string   `env:"ALLOWED_DOMAIN"`
	SearchUrls    []string `env:"SEARCH_URLS" envSeparator:","`
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load(".env.prod"); err != nil {
			log.Println("No .env file found")
		}
	}
}

func GetConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return &cfg
}
