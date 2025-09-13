package controller

import (
	"github.com/gin-gonic/gin"
	healthzService "github.com/skamranahmed/go-bank/internal/healthz/service"
)

type Dependency struct {
	HealthzService healthzService.HealthzService
}

func Register(router *gin.Engine, dependency Dependency) {
	healthzController := newHealthzController(dependency)
	router.GET("/healthz", healthzController.CheckHealth)
}
