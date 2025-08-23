package router

import (
	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/internal/healthz"
)

func Init(services *internal.Services) *gin.Engine {
	router := gin.Default()
	router.Use(gin.Recovery())

	healthz.Register(router, services.HealthzService)
	return router
}
