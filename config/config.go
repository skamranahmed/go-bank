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

const (
	APP_ENVIRONMENT_LOCAL      string = "local"
	APP_ENVIRONMENT_TEST       string = "test"
	APP_ENVIRONMENT_STAGING    string = "staging"
	APP_ENVIRONMENT_PRODUCTION string = "production"
)

func loadConfig() *Config {
	once.Do(func() {
		loadConfigFile("default.yaml")
		loadConfigFile("local.yaml")

		if GetEnvironment() == "staging" {
			loadConfigFile("staging.yaml")
		}

		if GetEnvironment() == "production" {
			loadConfigFile("production.yaml")
		}

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

func GetEnvironment() string {
	return getEnvironment()
}

func GetServerConfig() ServerConfig {
	serverConfig := loadConfig().Server

	serverPort := getServerPort()
	if serverPort != 0 {
		serverConfig.Port = serverPort
	}

	gracefulShutdownTimeout := getServerGracefulShutdownTimeoutInSeconds()
	if gracefulShutdownTimeout != 0 {
		serverConfig.GracefulShutdownTimeoutInSeconds = gracefulShutdownTimeout
	}

	return serverConfig
}

func GetPostgresConfig() PostgresConfig {
	postgresConfig := loadConfig().Database.Postgres

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

	caCertificate := getPostgresOptionsCaCertificate()
	if caCertificate != "" {
		postgresConfig.Options.CaCertificate = caCertificate
	}

	maxOpenConns := getPostgresOptionsMaxOpenConnections()
	if maxOpenConns != 0 {
		postgresConfig.Options.MaxOpenConnections = maxOpenConns
	}

	maxIdleConns := getPostgresOptionsMaxIdleConnections()
	if maxIdleConns != 0 {
		postgresConfig.Options.MaxIdleConnections = maxIdleConns
	}

	connectionMaxLifetimeInSeconds := getPostgresOptionsConnectionMaxLifetimeInSeconds()
	if connectionMaxLifetimeInSeconds != -1 {
		postgresConfig.Options.ConnectionMaxLifetimeInSeconds = connectionMaxLifetimeInSeconds
	}

	dialTimeoutInSeconds := getPostgresOptionsDialTimeoutInSeconds()
	if dialTimeoutInSeconds != -1 {
		postgresConfig.Options.DialTimeoutInSeconds = dialTimeoutInSeconds
	}

	readTimeoutInSeconds := getPostgresOptionsReadTimeoutInSeconds()
	if readTimeoutInSeconds != -1 {
		postgresConfig.Options.ReadTimeoutInSeconds = readTimeoutInSeconds
	}

	writeTimeoutInSeconds := getPostgresOptionsWriteTimeoutInSeconds()
	if writeTimeoutInSeconds != -1 {
		postgresConfig.Options.WriteTimeoutInSeconds = writeTimeoutInSeconds
	}

	return postgresConfig
}

func GetRedisConfig() RedisConfig {
	redisConfig := loadConfig().Cache.Redis

	host := getRedisHost()
	if host != "" {
		redisConfig.Host = host
	}

	port := getRedisPort()
	if port != 0 {
		redisConfig.Port = port
	}

	password := getRedisPassword()
	if password != "" {
		redisConfig.Password = password
	}

	dbIndex := getRedisDbIndex()
	if dbIndex != -1 {
		redisConfig.DbIndex = dbIndex
	}

	return redisConfig
}
