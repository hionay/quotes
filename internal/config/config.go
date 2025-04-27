package config

import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

const (
	envMySQLDSN       = "MYSQL_DSN"
	envServerPort     = "SERVER_PORT"
	envDBMaxOpenConns = "DB_MAX_OPEN_CONNS"
	envDBMaxIdleConns = "DB_MAX_IDLE_CONNS"
)

const (
	defaultServerPort     = 8080
	defaultMySQLPort      = 3306
	defaultDBMaxOpenConns = 3
)

type Config struct {
	opts Options
}

func NewConfig() *Config {
	opts := ReadOptionsFromEnv()
	return &Config{opts: opts}
}

func (c *Config) MySQLDSN() string {
	return c.opts.MySQLDSN
}

func (c *Config) ServerPort() int {
	return c.opts.ServerPort
}

func (c *Config) DBMaxOpenConns() int {
	return c.opts.DBMaxOpenConns
}

func (c *Config) DBMaxIdleConns() int {
	return c.opts.DBMaxIdleConns
}

type Options struct {
	MySQLDSN       string
	ServerPort     int
	DBMaxOpenConns int
	DBMaxIdleConns int
}

func ReadOptionsFromEnv() Options {
	return Options{
		MySQLDSN:       getEnvString(envMySQLDSN, ""),
		ServerPort:     getEnvInt(envServerPort, defaultServerPort),
		DBMaxOpenConns: getEnvInt(envDBMaxOpenConns, defaultDBMaxOpenConns),
		DBMaxIdleConns: getEnvInt(envDBMaxIdleConns, defaultDBMaxOpenConns),
	}
}

func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
