package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	healthzService "github.com/skamranahmed/go-bank/internal/healthz/service"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func Register(router *gin.Engine, service healthzService.HealthzService) {
	healthzController := newHealthzController(service)
	router.GET("/healthz", healthzController.CheckHealth)
	router.GET("/db-ping", healthzController.DbPing)
}

type healthzController struct {
	service healthzService.HealthzService
}

func newHealthzController(service healthzService.HealthzService) HealthzController {
	return &healthzController{
		service: service,
	}
}

func (c *healthzController) CheckHealth(ginCtx *gin.Context) {
	ginCtx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (c *healthzController) DbPing(ginCtx *gin.Context) {
	// check readiness
	err := c.service.DbPing()
	if err != nil {
		logger.Errorf("unable to connect to postgres db, error: %+v", err)
		ginCtx.JSON(http.StatusInternalServerError, gin.H{
			"status": "not ok",
		})
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
