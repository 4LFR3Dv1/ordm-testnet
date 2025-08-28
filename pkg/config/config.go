package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	Server struct {
		Port         string        `env:"PORT" default:"3000"`
		Host         string        `env:"HOST" default:"localhost"`
		ReadTimeout  time.Duration `env:"READ_TIMEOUT" default:"30s"`
		WriteTimeout time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
	}

	Auth struct {
		AdminUser     string        `env:"ADMIN_USER" default:"admin"`
		AdminPassword string        `env:"ADMIN_PASSWORD" required:"true"`
		JWTSecret     string        `env:"JWT_SECRET" required:"true"`
		SessionTTL    time.Duration `env:"SESSION_TTL" default:"24h"`
	}

	Security struct {
		RateLimitAttempts int           `env:"RATE_LIMIT_ATTEMPTS" default:"3"`
		RateLimitWindow   time.Duration `env:"RATE_LIMIT_WINDOW" default:"5m"`
		LockoutDuration   time.Duration `env:"LOCKOUT_DURATION" default:"15m"`
		CSRFSecret        string        `env:"CSRF_SECRET" required:"true"`
	}
}

var AppConfig Config

func LoadConfig() error {
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		log.Fatal("ADMIN_PASSWORD environment variable is required")
	}
	AppConfig.Auth.AdminPassword = adminPass

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		AppConfig.Auth.JWTSecret = jwtSecret
	} else {
		AppConfig.Auth.JWTSecret = "ordm-jwt-secret-dev"
	}

	if csrfSecret := os.Getenv("CSRF_SECRET"); csrfSecret != "" {
		AppConfig.Security.CSRFSecret = csrfSecret
	} else {
		AppConfig.Security.CSRFSecret = "ordm-csrf-secret-dev"
	}

	return nil
}

func GetPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "3000"
}

func IsProduction() bool {
	return os.Getenv("ENV") == "production"
}
