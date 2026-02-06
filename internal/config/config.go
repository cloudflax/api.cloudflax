package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config contiene la configuración de la aplicación.
type Config struct {
	Port string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
}

// Load carga la configuración desde variables de entorno.
// En Docker, las variables se configuran via env_file o environment en docker-compose.
func Load() (*Config, error) {
	cfg := &Config{
		Port: getEnv("PORT", "3000"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "cloudflax"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate verifica que la configuración obligatoria esté presente.
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT es obligatorio")
	}
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST es obligatorio")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER es obligatorio")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME es obligatorio")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}
