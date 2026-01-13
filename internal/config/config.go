package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type JWTConfig struct {
	Secret string `yaml:"secret" env:"JWT_SECRET" envDefault:"your-very-random-secret-key-here"`
	TTL    string `yaml:"ttl" env:"JWT_TTL" envDefault:"15m"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	DBName   string `yaml:"dbname" env:"POSTGRES_DB" envDefault:"mydb"`
	SSLMode  string `yaml:"sslmode" env:"POSTGRES_SSLMODE" envDefault:"disable"`
}

type RedisConfig struct {
	Host     string `yaml:"host" env:"REDIS_HOST" envDefault:"localhost"`
	Port     int    `yaml:"port" env:"REDIS_PORT" envDefault:"6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `yaml:"db" env:"REDIS_DB" envDefault:"0"`
	TTL      string `yaml:"ttl" env:"REDIS_TTL" envDefault:"300s"`
}

type Config struct {
	Env        string         `yaml:"env" env:"ENV" envDefault:"local" envRequired:"true"`
	Postgres   PostgresConfig `yaml:"postgres"`
	Redis      RedisConfig    `yaml:"redis"`
	HTTPServer struct {
		Address     string `yaml:"address" env:"HTTP_ADDRESS" envDefault:"localhost:8080"`
		Timeout     string `yaml:"timeout" env:"HTTP_TIMEOUT" envDefault:"4s"`
		IdleTimeout string `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	} `yaml:"http_server"`
	JWT JWTConfig `yaml:"jwt"`
}

func (r *RedisConfig) TTLDuration() time.Duration {
	d, err := time.ParseDuration(r.TTL)
	if err != nil {
		log.Fatalf("invalid redis TTL duration: %s", err)
	}
	return d
}

func (c *Config) JWTTTLDuration() time.Duration {
	d, err := time.ParseDuration(c.JWT.TTL)
	if err != nil {
		log.Fatalf("invalid JWT TTL duration: %s", err)
	}
	return d
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
