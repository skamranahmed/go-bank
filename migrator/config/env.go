package config

import (
	"os"
	"strconv"
)

const (
	postgresHost         = "POSTGRES_HOST"
	postgresUsername     = "POSTGRES_USERNAME"
	postgresPassword     = "POSTGRES_PASSWORD"
	postgresDatabaseName = "POSTGRES_DATABASE_NAME"
	postgresPort         = "POSTGRES_PORT"
)

func getPostgresHost() string {
	return os.Getenv(postgresHost)
}

func getPostgresUsername() string {
	return os.Getenv(postgresUsername)
}

func getPostgresPassword() string {
	return os.Getenv(postgresPassword)
}

func getPostgresDatabaseName() string {
	return os.Getenv(postgresDatabaseName)
}

func getPostgresPort() int {
	portNumber, err := strconv.Atoi(os.Getenv(postgresPort))
	if err != nil {
		return 0
	}
	return portNumber
}
