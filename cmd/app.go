package cmd

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/skamranahmed/go-bank/cmd/router"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/cmd/worker"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/skamranahmed/go-bank/pkg/database"
	"github.com/skamranahmed/go-bank/pkg/logger"
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

	asyncqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DbIndex,
	})
	defer asyncqClient.Close()

	services, err := internal.BootstrapServices(db, cacheClient, asyncqClient)
	if err != nil {
		return err
	}

	if role == RoleServer {
		router := router.Init(db, services)
		server.Start(router)
	}

	if role == RoleWorkerDefault {
		// start worker that will consume tasks from the "default" queue
		worker.Start(worker.DefaultQueue, services)
	}

	if role == RoleWorkerPriority {
		// start worker that will consume tasks from the "priority" queue
		worker.Start(worker.PriorityQueue, services)
	}

	return nil
}
