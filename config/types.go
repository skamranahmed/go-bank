package config

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
)

type Config struct {
	Environment string          `koanf:"environment"`
	Logger      LoggerConfig    `koanf:"logger"`
	Server      ServerConfig    `koanf:"server"`
	Telemetry   TelemetryConfig `koanf:"telemetry"`
	Database    DatabaseConfig  `koanf:"database"`
	Cache       CacheConfig     `koanf:"cache"`
	Auth        AuthConfig      `koanf:"auth"`
}

type LoggerConfig struct {
	Level LogLevel `koanf:"level"`
}

type ServerConfig struct {
	Port                             int `koanf:"port"`
	GracefulShutdownTimeoutInSeconds int `koanf:"gracefulShutdownTimeoutInSeconds"`
}

type TelemetryConfig struct {
	ServiceName          string `koanf:"serviceName"`
	TracesIntakeEndpoint string `koanf:"tracesIntakeEndpoint"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `koanf:"postgres"`
}

type PostgresConfig struct {
	Host         string `koanf:"host"`
	Username     string `koanf:"username"`
	Password     string `koanf:"password"`
	DatabaseName string `koanf:"databaseName"`
	Port         int    `koanf:"port"`
	Options      struct {
		CaCertificate                  string `koanf:"caCertificate"`
		MaxOpenConnections             int    `koanf:"maxOpenConnections"`
		MaxIdleConnections             int    `koanf:"maxIdleConnections"`
		ConnectionMaxLifetimeInSeconds int    `koanf:"connectionMaxLifetimeInSeconds"`
		DialTimeoutInSeconds           int    `koanf:"dialTimeoutInSeconds"`
		ReadTimeoutInSeconds           int    `koanf:"readTimeoutInSeconds"`
		WriteTimeoutInSeconds          int    `koanf:"writeTimeoutInSeconds"`
	} `koanf:"options"`
}

type CacheConfig struct {
	Redis RedisConfig `koanf:"redis"`
}

type RedisConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DbIndex  int    `koanf:"dbIndex"`
}

type AuthConfig struct {
	AccessTokenExpiryDurationInSeconds int    `koanf:"accessTokenExpiryDurationInSeconds"`
	AccessTokenSecretSigningKey        string `koanf:"accessTokenSecretSigningKey"`
}
