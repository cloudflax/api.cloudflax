package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config contiene la configuración de la aplicación.
type Config struct {
	Port string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	AWSAccessKey string
	AWSSecretKey string
	AWSS3Bucket  string
}

// Load carga la configuración desde .env y variables de entorno.
// Si existe .env, lo carga; si no, usa solo las env vars (útil en Docker).
func Load() (*Config, error) {
	_ = godotenv.Load() // Ignora error si .env no existe

	cfg := &Config{
		Port: getEnv("PORT", "3000"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "cloudflax"),

		AWSAccessKey: getEnv("AWS_ACCESS_KEY", ""),
		AWSSecretKey: getEnv("AWS_SECRET_KEY", ""),
		AWSS3Bucket:   getEnv("AWS_S3_BUCKET", ""),
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
