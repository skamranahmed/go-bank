package config

import (
	"os"
	"strconv"
)

const (
	// logger
	loggerLevel = "LOGGER_LEVEL"

	environment = "ENVIRONMENT"

	// server
	serverPort                             = "SERVER_PORT"
	serverGracefulShutdownTimeoutInSeconds = "SERVER_GRACEFUL_SHUTDOWN_TIMEOUT_IN_SECONDS"

	// postgres
	postgresHost                                  = "POSTGRES_HOST"
	postgresUsername                              = "POSTGRES_USERNAME"
	postgresPassword                              = "POSTGRES_PASSWORD"
	postgresDatabaseName                          = "POSTGRES_DATABASE_NAME"
	postgresPort                                  = "POSTGRES_PORT"
	postgresOptionsCaCertificate                  = "POSTGRES_OPTIONS_CA_CERTIFICATE"
	postgresOptionsMaxOpenConnections             = "POSTGRES_OPTIONS_MAX_OPEN_CONNECTIONS"
	postgresOptionsMaxIdleConnections             = "POSTGRES_OPTIONS_MAX_IDLE_CONNECTIONS"
	postgresOptionsConnectionMaxLifetimeInSeconds = "POSTGRES_OPTIONS_CONNECTION_MAX_LIFETIME_IN_SECONDS"
	postgresOptionsDialTimeoutInSeconds           = "POSTGRES_OPTIONS_DIAL_TIMEOUT_IN_SECONDS"
	postgresOptionsReadTimeoutInSeconds           = "POSTGRES_OPTIONS_READ_TIMEOUT_IN_SECONDS"
	postgresOptionsWriteTimeoutInSeconds          = "POSTGRES_OPTIONS_WRITE_TIMEOUT_IN_SECONDS"

	// redis
	redisHost     = "REDIS_HOST"
	redisPort     = "REDIS_PORT"
	redisPassword = "REDIS_PASSWORD"
	redisDbIndex  = "REDIS_DB_INDEX"

	// auth
	authAccessTokenExpiryDurationInSeconds = "AUTH_ACCESS_TOKEN_EXPIRY_DURATION_IN_SECONDS"
	authAccessTokenSecretSigningKey        = "AUTH_ACCESS_TOKEN_SECRET_SIGNING_KEY"
)

func getLoggerLevel() string {
	return os.Getenv(loggerLevel)
}

func getEnvironment() string {
	env := os.Getenv(environment)
	if env == "" {
		return APP_ENVIRONMENT_LOCAL
	}
	return env
}

func getServerPort() int {
	portNumber, err := strconv.Atoi(os.Getenv(serverPort))
	if err != nil {
		return 0
	}
	return portNumber
}

func getServerGracefulShutdownTimeoutInSeconds() int {
	timeout, err := strconv.Atoi(os.Getenv(serverGracefulShutdownTimeoutInSeconds))
	if err != nil {
		return 0
	}
	return timeout
}

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

func getPostgresOptionsCaCertificate() string {
	return os.Getenv(postgresOptionsCaCertificate)
}

func getPostgresOptionsMaxOpenConnections() int {
	maxOpenConnections, err := strconv.Atoi(os.Getenv(postgresOptionsMaxOpenConnections))
	if err != nil {
		return 0
	}
	return maxOpenConnections
}

func getPostgresOptionsMaxIdleConnections() int {
	maxIdleConnections, err := strconv.Atoi(os.Getenv(postgresOptionsMaxIdleConnections))
	if err != nil {
		return 0
	}
	return maxIdleConnections
}

func getPostgresOptionsConnectionMaxLifetimeInSeconds() int {
	maxLifetime, err := strconv.Atoi(os.Getenv(postgresOptionsConnectionMaxLifetimeInSeconds))
	if err != nil {
		// since 0 is a valid value for connection max lifetime in postgres
		// to indicate that an error has occured, we are returning -1
		return -1
	}
	return maxLifetime
}

func getPostgresOptionsDialTimeoutInSeconds() int {
	dialTimeout, err := strconv.Atoi(os.Getenv(postgresOptionsDialTimeoutInSeconds))
	if err != nil {
		// since 0 is a valid value for dial timeout in postgres
		// to indicate that an error has occured, we are returning -1
		return -1
	}
	return dialTimeout
}

func getPostgresOptionsReadTimeoutInSeconds() int {
	readTimeout, err := strconv.Atoi(os.Getenv(postgresOptionsReadTimeoutInSeconds))
	if err != nil {
		// since 0 is a valid value for read timeout in postgres
		// to indicate that an error has occured, we are returning -1
		return -1
	}
	return readTimeout
}

func getPostgresOptionsWriteTimeoutInSeconds() int {
	writeTimeout, err := strconv.Atoi(os.Getenv(postgresOptionsWriteTimeoutInSeconds))
	if err != nil {
		// since 0 is a valid value for read timeout in postgres
		// to indicate that an error has occured, we are returning -1
		return -1
	}
	return writeTimeout
}

func getRedisHost() string {
	return os.Getenv(redisHost)
}

func getRedisPort() int {
	portNumber, err := strconv.Atoi(os.Getenv(redisPort))
	if err != nil {
		return 0
	}
	return portNumber
}

func getRedisPassword() string {
	return os.Getenv(redisPassword)
}

func getRedisDbIndex() int {
	dbIndex, err := strconv.Atoi(os.Getenv(redisDbIndex))
	if err != nil {
		// since 0 is a valid redis db index, to indicate that an error has occured, we are returning -1
		return -1
	}
	return dbIndex
}

func getAccessTokenExpiryDurationInSeconds() int {
	duration, err := strconv.Atoi(os.Getenv(authAccessTokenExpiryDurationInSeconds))
	if err != nil {
		// since 0 is a valid expiry duration, to indicate that an error has occured, we are returning -1
		return -1
	}
	return duration
}

func getAccessTokenSecretSigningKey() string {
	return os.Getenv(authAccessTokenSecretSigningKey)
}
