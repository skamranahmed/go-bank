package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"time"

	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"
)

func NewPostgresClient() (*bun.DB, error) {
	postgresConfig := config.GetPostgresConfig()

	pgDriverOptions := []pgdriver.Option{
		pgdriver.WithAddr(fmt.Sprintf("%s:%d", postgresConfig.Host, postgresConfig.Port)),
		pgdriver.WithUser(postgresConfig.Username),
		pgdriver.WithPassword(postgresConfig.Password),
		pgdriver.WithDatabase(postgresConfig.DatabaseName),
		pgdriver.WithDialTimeout(time.Duration(postgresConfig.Options.DialTimeoutInSeconds) * time.Second),
		pgdriver.WithReadTimeout(time.Duration(postgresConfig.Options.ReadTimeoutInSeconds) * time.Second),
		pgdriver.WithWriteTimeout(time.Duration(postgresConfig.Options.WriteTimeoutInSeconds) * time.Second),
	}

	if config.GetEnvironment() == config.APP_ENVIRONMENT_LOCAL || config.GetEnvironment() == config.APP_ENVIRONMENT_TEST {
		// we can use an insecure connection
		pgDriverOptions = append(pgDriverOptions, pgdriver.WithInsecure(true))
	} else {
		// we must use a secure connection
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM([]byte(postgresConfig.Options.CaCertificate)) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		pgDriverOptions = append(pgDriverOptions, pgdriver.WithTLSConfig(&tls.Config{
			RootCAs:    certPool,
			ServerName: postgresConfig.Host,
		}))
	}

	sqlDb := sql.OpenDB(pgdriver.NewConnector(pgDriverOptions...))

	db := bun.NewDB(sqlDb, pgdialect.New())

	// check readiness
	err := db.Ping()
	if err != nil {
		logger.Error(context.TODO(), "Unable to connect to postgres db, error: %+v", err)
		return nil, err
	}

	if config.GetLoggerConfig().Level == config.LogLevelDebug {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true), // print full SQL with args
		))
	}

	// enable opentelemetry tracing
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName("postgres")))

	// connection pooling
	db.SetMaxOpenConns(postgresConfig.Options.MaxOpenConnections)
	db.SetMaxIdleConns(postgresConfig.Options.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(postgresConfig.Options.ConnectionMaxLifetimeInSeconds) * time.Second)

	return db, nil
}
