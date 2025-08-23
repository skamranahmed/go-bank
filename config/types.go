package config

type Config struct {
	Environment string         `koanf:"environment"`
	Server      ServerConfig   `koanf:"server"`
	Database    DatabaseConfig `koanf:"database"`
	Cache       CacheConfig    `koanf:"cache"`
}

type ServerConfig struct {
	Port                             int `koanf:"port"`
	GracefulShutdownTimeoutInSeconds int `koanf:"gracefulShutdownTimeoutInSeconds"`
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
