package service

import (
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/uptrace/bun"
)

type healthzService struct {
	dbClient    *bun.DB
	cacheClient cache.CacheClient
}

func NewHealthzService(dbClient *bun.DB, cacheClient cache.CacheClient) HealthzService {
	return &healthzService{
		dbClient:    dbClient,
		cacheClient: cacheClient,
	}
}

func (h *healthzService) DbPing() error {
	return h.dbClient.Ping()
}

func (h *healthzService) CachePing() error {
	return  h.cacheClient.Ping()
}