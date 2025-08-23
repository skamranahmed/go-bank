package internal

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal/healthz"
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/skamranahmed/go-bank/pkg/database"
)

type Services struct {
	HealthzService healthz.HealthzService
}

func BootstrapServices() (*Services, error) {
	// initialize postgres
	db, err := database.NewPostgresClient()
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer cacheClient.Close()

	healthzService := healthz.NewHealthzService(db, cacheClient)

	return &Services{
		HealthzService: healthzService,
	}, nil
}
