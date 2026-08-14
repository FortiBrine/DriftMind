package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

func (e Environment) IsDev() bool { return e == EnvDev }

type Config struct {
	Environment     Environment   `env:"ENVIRONMENT" envDefault:"dev"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`

	TelegramToken string `env:"TELEGRAM_TOKEN,required"`
	WebhookURL    string `env:"WEBHOOK_URL"`
	HttpAddress   string `env:"HTTP_ADDRESS" envDefault:":2000"`
}

func Load() (Config, error) {
	godotenv.Load()
	return env.ParseAs[Config]()
}
