package cmd

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/skamranahmed/go-bank/cmd/router"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/cmd/worker"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/skamranahmed/go-bank/pkg/database"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
	"github.com/skamranahmed/go-bank/pkg/telemetry"
)

const (
	RoleServer         string = "server"
	RoleWorkerDefault  string = "worker-default"
	RoleWorkerPriority string = "worker-priority"
)

var validRoles = map[string]struct{}{
	RoleServer:         {},
	RoleWorkerDefault:  {},
	RoleWorkerPriority: {},
}

func Run(role string) error {
	_, ok := validRoles[role]
	if !ok {
		return fmt.Errorf("Unsupported role provided for server startup, role: %+v", role)
	}

	config.Role = role

	logger.Init()

	// initialize postgres
	db, err := database.NewPostgresClient()
	if err != nil {
		return err
	}
	defer db.Close()

	// initialize redis
	redisConfig := config.GetRedisConfig()
	cacheClient, err := cache.NewRedisClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DbIndex,
	})
	if err != nil {
		return err
	}
	defer cacheClient.Close()

	taskEnqueuer := tasksHelper.NewAsynqTaskEnqueuer(tasksHelper.AsynqRedisConfig{
		Host:     redisConfig.Host,
		Port:     redisConfig.Port,
		Password: redisConfig.Password,
		DbIndex:  redisConfig.DbIndex,
	})
	defer taskEnqueuer.Close()

	services, err := internal.BootstrapServices(db, cacheClient, taskEnqueuer)
	if err != nil {
		return err
	}

	// initialize OpenTelemetry tracer
	ctx := context.TODO()
	tracerProvider, err := telemetry.InitTracer()
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize otel tracer provider: %+v", err)
	}
	defer tracerProvider.Shutdown(ctx)

	if role == RoleServer {
		router := router.Init(db, services)
		server.Start(router)
	}

	if role == RoleWorkerDefault {
		// start worker that will consume tasks from the "default" queue
		worker.Start(tasksHelper.DefaultQueue, services)
	}

	if role == RoleWorkerPriority {
		// start worker that will consume tasks from the "priority" queue
		worker.Start(tasksHelper.PriorityQueue, services)
	}

	return nil
}
