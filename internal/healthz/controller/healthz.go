package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	healthzService "github.com/skamranahmed/go-bank/internal/healthz/service"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

type healthzController struct {
	service healthzService.HealthzService
}

func newHealthzController(dependency Dependency) HealthzController {
	return &healthzController{
		service: dependency.HealthzService,
	}
}

func (c *healthzController) CheckHealth(ginCtx *gin.Context) {
	// check db readiness
	err := c.service.DbPing()
	if err != nil {
		logger.Errorf("unable to connect to postgres db, error: %+v", err)
		ginCtx.JSON(http.StatusInternalServerError, gin.H{
			"status": "DB_NOT_OK",
		})
		return
	}

	// check cache readiness
	err = c.service.CachePing()
	if err != nil {
		logger.Errorf("unable to connect to redis, error: %+v", err)
		ginCtx.JSON(http.StatusInternalServerError, gin.H{
			"status": "CACHE_NOT_OK",
		})
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{
		"status": "ALL_OK",
	})
}
