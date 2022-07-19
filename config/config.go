package config

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig AppConfig
	DB        DatabaseConfig
	Reload    Reload
}

type AppConfig struct {
	Name    string `env:"CONFIG__APP_CONFIG__NAME" required:"true" default:"card-delivery-service"`
	Version string `env:"APP__VERSION" default:"local"`
	Port    int    `env:"CONFIG__APP_CONFIG__PORT" default:"3000"`
}

type Reload struct {
	Trigger string `env:"CONFIG__TRIGGER" default:"default"`
}

type DatabaseConfig struct {
	Host     string `env:"CONFIG__DB__HOST" required:"true"`
	Port     int    `env:"CONFIG__DB__PORT" default:"3306"`
	User     string `env:"CONFIG__DB__USER" required:"true"`
	Password string `env:"CONFIG__DB__PASSWORD" required:"true"`
	Name     string `env:"CONFIG__DB__NAME" required:"true"`
}

func LoadConfig() (*Config, error) {
	var config Config
	err := configor.
		New(&configor.Config{AutoReload: false}).
		Load(&config, fmt.Sprintf("%s/config.%s.json", getConfigLocation(), getEnv()))

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getConfigLocation() string {
	_, filename, _, _ := runtime.Caller(0)

	return path.Join(path.Dir(filename), "../config")
}

func getEnv() string {
	val := os.Getenv("APP_ENV")
	// todo: check our stage names and align with them
	switch val {
	case "prod":
		return "prod"
	case "test":
		return "test"
	case "qa":
		return "qa"
	default:
		return "dev"
	}
}
