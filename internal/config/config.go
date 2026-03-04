package config

import (
	"github.com/wb-go/wbf/config"
)

// Config — структура конфигурации приложения
type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
}

// Server — конфигурация сервера
type Server struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

// Database — конфигурация базы данных
type Database struct {
	DSN string `mapstructure:"dsn"`
}

// Load загружает конфигурацию
func Load(path string) (*Config, error) {
	cfg := config.New()

	// Загрузка .env файла
	_ = cfg.LoadEnvFiles(".env")

	// Включение переменных окружения
	cfg.EnableEnv("")

	// Загрузка конфигурационного файла
	if err := cfg.LoadConfigFiles(path); err != nil {
		return nil, err
	}

	var appCfg Config
	if err := cfg.UnmarshalExact(&appCfg); err != nil {
		return nil, err
	}

	return &appCfg, nil
}
