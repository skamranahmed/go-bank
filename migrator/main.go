package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/skamranahmed/go-bank/migrator/config"
	_ "github.com/skamranahmed/go-bank/migrator/migrations"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Please provide a Goose command")
	}

	postgresConfig := config.GetPostgresConfig()

	dbDsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		postgresConfig.Username,
		postgresConfig.Password,
		postgresConfig.DatabaseName, postgresConfig.Host,
		postgresConfig.Port,
	)

	db, err := sql.Open("postgres", dbDsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %+v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("unable to establish connection with the db, error: %+v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %+v", err)
	}

	goose.SetTableName("goose_migrations_metadata")

	command := args[0]
	migrationDir := "migrations"

	if command == "up" {
		err = goose.UpContext(context.Background(), db, migrationDir, goose.WithAllowMissing())
		if err != nil {
			if errors.Is(err, goose.ErrNoMigrationFiles) {
				// capturing the 'ErrNoMigrationFiles' error else it will lead to a non-zero exit code
				fmt.Println("No pending migration files found. Skipping 'up' migrations")
			} else {
				log.Fatalf("Failed to run Goose command: %+v", err)
			}
		}
	} else {
		err = goose.RunContext(context.Background(), command, db, migrationDir, args[1:]...)
		if err != nil {
			if errors.Is(err, goose.ErrNoCurrentVersion) {
				// capturing the 'ErrNoCurrentVersion' error else it will lead to a non-zero exit code
				fmt.Println("No applied migrations available to rollback. Skipping 'down' migrations")
			} else {
				log.Fatalf("Failed to run Goose command: %+v", err)
			}
		}
	}

}
