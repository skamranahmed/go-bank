package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/skamranahmed/go-bank/cmd/router"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	accountModel "github.com/skamranahmed/go-bank/internal/account/model"
	userModel "github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type PostgresTestContainer struct {
	MappedPort   string
	TeardownFunc func()
}

type RedisTestContainer struct {
	MappedPort   string
	TeardownFunc func()
}

var (
	postgresTestContainerOnce sync.Once
	postgresTestContainer     PostgresTestContainer

	redisTestContainerOnce sync.Once
	redisTestContainer     RedisTestContainer
)

type TestApp struct {
	Db           *bun.DB
	Cache        cache.CacheClient
	Services     *internal.Services
	Router       *gin.Engine
	TeardownFunc func()
}

type TestAppDeps struct {
	Db           *bun.DB
	Cache        cache.CacheClient
	TaskEnqueuer tasksHelper.TaskEnqueuer
}

// NewTestApp creates a TestApp for testing.
// Supports optional dependency injection via deps
func NewTestApp(ctx context.Context, deps *TestAppDeps, postresTestContainer *PostgresTestContainer, redisTestContainer *RedisTestContainer) TestApp {
	var db *bun.DB
	var cache cache.CacheClient
	var taskEnqueuer tasksHelper.TaskEnqueuer

	if deps != nil {
		// use provided dependencies if available
		if deps.TaskEnqueuer != nil {
			taskEnqueuer = deps.TaskEnqueuer
		}

		if deps.Db != nil {
			db = deps.Db
		}

		if deps.Cache != nil {
			cache = deps.Cache
		}
	}

	// fallback to real implementations if nil
	if taskEnqueuer == nil {
		taskEnqueuer = asynq.NewClient(asynq.RedisClientOpt{
			Addr: fmt.Sprintf("localhost:%s", redisTestContainer.MappedPort),
		})
	}

	if db == nil {
		db = setupPostgresDb(ctx, postresTestContainer)
	}

	if cache == nil {
		cache = setupRedis(ctx, redisTestContainer)
	}

	services, err := internal.BootstrapServices(db, cache, taskEnqueuer)
	if err != nil {
		logger.Fatal(ctx, "Unable to bootstrap services, error: %+v", err)
	}
	testRouter := router.Init(db, services)

	return TestApp{
		Db:       db,
		Cache:    cache,
		Services: services,
		Router:   testRouter,
		TeardownFunc: func() {
			db.Close()
			cache.Close()
		},
	}
}

func NewPostgresTestContainer(ctx context.Context) *PostgresTestContainer {
	postgresTestContainerOnce.Do(func() {
		const postgresTestDBName string = "go_bank_test"
		const postgresTestDBPassword string = "go_bank_test"
		const postgresTestDBUser string = "go_bank_test"

		containerReq := testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
			Env: map[string]string{
				"POSTGRES_DB":       postgresTestDBName,
				"POSTGRES_PASSWORD": postgresTestDBPassword,
				"POSTGRES_USER":     postgresTestDBUser,
			},
		}

		dbContainer, err := testcontainers.GenericContainer(
			ctx,
			testcontainers.GenericContainerRequest{
				ContainerRequest: containerReq,
				Started:          true,
			})
		if err != nil {
			logger.Fatal(ctx, "Unable to start db container, error: %+v", err)
		}

		port, err := dbContainer.MappedPort(ctx, "5432")
		if err != nil {
			logger.Fatal(ctx, "Unable to get port for db container, error: %+v", err)
		}

		postgresTestContainer = PostgresTestContainer{
			MappedPort: port.Port(),
			TeardownFunc: func() {
				err := dbContainer.Terminate(ctx)
				if err != nil {
					logger.Fatal(ctx, err.Error())
				}
			},
		}
	})

	return &postgresTestContainer
}

func setupPostgresDb(ctx context.Context, postgresTestContainer *PostgresTestContainer) *bun.DB {
	defaultPostgresDbDsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		"go_bank_test",
		"go_bank_test",
		fmt.Sprintf("localhost:%s", postgresTestContainer.MappedPort),
		"postgres",
	)
	sqlDbForOperations := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(defaultPostgresDbDsn)))
	operationsDb := bun.NewDB(sqlDbForOperations, pgdialect.New())
	if config.GetLoggerConfig().Level == config.LogLevelDebug {
		operationsDb.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true), // print full SQL with args
		))
	}

	// make db template if doesn't exist
	// check if template exists or not
	var doesDbTemplateExist bool
	err := operationsDb.QueryRowContext(ctx, `
		SELECT
			EXISTS (
				SELECT
					1
				FROM
					pg_database
				WHERE
					datname = 'go_bank_test_template'
				);
		`).Scan(&doesDbTemplateExist)
	if err != nil {
		logger.Fatal(ctx, "Unable to check for template database existence, error: %+v", err)
	}

	// if it does not exist, bootstrap the db and then create a template out of it
	if !doesDbTemplateExist {
		// prepare the db by performing operations on db such as defining the table schemas etc...
		for _, model := range allModels() {
			_, err := operationsDb.NewDropTable().IfExists().Cascade().Model(model).Exec(ctx)
			if err != nil {
				logger.Fatal(ctx, err.Error())
			}

			_, err = operationsDb.NewCreateTable().Model(model).WithForeignKeys().Exec(ctx)
			if err != nil {
				logger.Fatal(ctx, err.Error())
			}

		}

		_, err := operationsDb.NewRaw(`CREATE DATABASE "go_bank_test_template" TEMPLATE "postgres"`).Exec(ctx)
		if err != nil {
			logger.Fatal(ctx, "Unable to create database `go_bank_test_template`, error: %+v", err)
		}
	}

	// ensure fresh database
	// switch to a different db connection before doing
	_, err = operationsDb.NewRaw(`DROP DATABASE "go_bank_test"`).Exec(ctx)
	if err != nil {
		logger.Fatal(ctx, "Unable to drop database `go_bank_test`, error: %+v", err)
	}

	_, err = operationsDb.NewRaw(`CREATE DATABASE "go_bank_test" TEMPLATE "go_bank_test_template"`).Exec(ctx)
	if err != nil {
		logger.Fatal(ctx, "Unable to create database `go_bank_test`, error: %+v", err)
	}

	testPostgresDbDsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		"go_bank_test",
		"go_bank_test",
		fmt.Sprintf("localhost:%s", postgresTestContainer.MappedPort),
		"go_bank_test",
	)
	sqlDbForTest := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(testPostgresDbDsn)))
	testDb := bun.NewDB(sqlDbForTest, pgdialect.New())

	if config.GetLoggerConfig().Level == config.LogLevelDebug {
		testDb.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true), // print full SQL with args
		))
	}

	return testDb
}

func NewRedisTestContainer(ctx context.Context) *RedisTestContainer {
	redisTestContainerOnce.Do(func() {
		containerReq := testcontainers.ContainerRequest{
			Image:        "redis:7.4.0-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		}

		redisContainer, err := testcontainers.GenericContainer(
			ctx,
			testcontainers.GenericContainerRequest{
				ContainerRequest: containerReq,
				Started:          true,
			})
		if err != nil {
			logger.Fatal(ctx, "Unable to start redis container, error: %+v", err)
		}

		port, err := redisContainer.MappedPort(ctx, "6379")
		if err != nil {
			logger.Fatal(ctx, "Unable to get port for redis container, error: %+v", err)
		}

		redisTestContainer = RedisTestContainer{
			MappedPort: port.Port(),
			TeardownFunc: func() {
				err := redisContainer.Terminate(ctx)
				if err != nil {
					logger.Fatal(ctx, err.Error())
				}
			},
		}
	})

	return &redisTestContainer
}

func setupRedis(ctx context.Context, redisTestContainer *RedisTestContainer) cache.CacheClient {
	ensureFreshRedis(ctx, redisTestContainer)

	redisClientOpts := &redis.Options{
		Addr: fmt.Sprintf("localhost:%s", redisTestContainer.MappedPort),
	}

	cacheClient, err := cache.NewRedisClient(redisClientOpts)
	if err != nil {
		logger.Fatal(ctx, "Redis test container error: %+v", err.Error())
	}

	return cacheClient
}

func ensureFreshRedis(ctx context.Context, redisTestContainer *RedisTestContainer) {
	redisClientOpts := &redis.Options{
		Addr: fmt.Sprintf("localhost:%s", redisTestContainer.MappedPort),
	}

	client := redis.NewClient(redisClientOpts)
	err := client.Ping(ctx).Err()
	if err != nil {
		logger.Fatal(ctx, "Unable to connect to redis, error: %+v", err)
	}

	_, err = client.FlushAll(ctx).Result()
	if err != nil {
		logger.Fatal(ctx, "Unable to flush redis keys, error: %+v", err)
	}
	client.Conn().Close()
}

func allModels() []interface{} {
	// must be in order so that any constraints and integrity checks are maintained
	return []interface{}{
		(*userModel.User)(nil),
		(*accountModel.Account)(nil),
		// add new models here
	}
}
