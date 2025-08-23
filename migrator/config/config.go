package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	k      = koanf.New(".")
	once   sync.Once
	config Config
)

func loadConfig() *Config {
	once.Do(func() {
		loadConfigFile("default.yaml")
		err := k.Unmarshal("", &config)
		if err != nil {
			log.Fatalf("error unmarshalling config: %v", err)
		}
	})

	return &config
}

func loadConfigFile(configFileName string) {
	dir, _ := os.Getwd()
	configFilPath := filepath.Join(dir, filepath.Join("config", "files", configFileName))
	k.Load(file.Provider(configFilPath), yaml.Parser())
}

func GetPostgresConfig() PostgresConfig {
	postgresConfig := loadConfig().Postgres

	host := getPostgresHost()
	if host != "" {
		postgresConfig.Host = host
	}

	username := getPostgresUsername()
	if username != "" {
		postgresConfig.Username = username
	}

	password := getPostgresPassword()
	if password != "" {
		postgresConfig.Password = password
	}

	dbName := getPostgresDatabaseName()
	if dbName != "" {
		postgresConfig.DatabaseName = dbName
	}

	port := getPostgresPort()
	if port != 0 {
		postgresConfig.Port = port
	}

	return postgresConfig
}
