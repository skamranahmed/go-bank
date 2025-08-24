package cmd

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/skamranahmed/go-bank/cmd/router"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/skamranahmed/go-bank/pkg/database"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func Run() error {
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

	services, err := internal.BootstrapServices(db, cacheClient)
	if err != nil {
		return err
	}

	router := router.Init(db, services)
	server.Start(router)

	return nil
}
