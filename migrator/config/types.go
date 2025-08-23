package config

type Config struct {
	Environment string         `koanf:"environment"`
	Postgres    PostgresConfig `koanf:"postgres"`
}

type PostgresConfig struct {
	Host         string `koanf:"host"`
	Username     string `koanf:"username"`
	Password     string `koanf:"password"`
	DatabaseName string `koanf:"databaseName"`
	Port         int    `koanf:"port"`
}
