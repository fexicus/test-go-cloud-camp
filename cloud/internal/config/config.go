package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Address             string        `yaml:"address" env-default:"localhost:8080"`
	Backends            []string      `yaml:"backends" env-required:"true"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval" env:"HEALTH_CHECK_INTERVAL" env-default:"10s"`
	Database            `yaml:"database"`
	DSN                 string `yaml:"dsn"`

	RateLimitCapacity int `yaml:"rate_limit_capacity" env-default:"10"`
	RateLimitRate     int `yaml:"rate_limit_rate"   env-default:"1"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"dbname"`
}

// GetDSN возвращает строку подключения или ошибку
func (c *Config) GetDSN() (string, error) {
	if c.DSN == "" {
		return "", fmt.Errorf("DSN не задан в конфигурации")
	}
	return c.DSN, nil
}

// MustLoad читает конфиг из файла, путь к которому в ENV CONFIG_PATH
func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		log.Fatal("ОШИБКА: переменная окружения CONFIG_PATH не задана")
	}
	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		log.Fatalf("ОШИБКА: не удалось прочитать конфигурационный файл: %v", err)
	}
	return cfg
}
