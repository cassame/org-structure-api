package config

import "os"

type Config struct {
	DBDSN string
	Port  string
}

func Load() *Config {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=user password=password dbname=org_db port=5432 sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DBDSN: dsn,
		Port:  port,
	}
}
