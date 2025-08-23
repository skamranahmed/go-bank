package healthz

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func Register(router *gin.Engine, service HealthzService) {
	healthzController := newHealthzController(service)
	router.GET("/healthz", healthzController.CheckHealth)
	router.GET("/db-ping", healthzController.DbPing)
}

type healthzController struct {
	service HealthzService
}

func newHealthzController(service HealthzService) HealthzController {
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
