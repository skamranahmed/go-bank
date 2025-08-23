package healthz

import "github.com/gin-gonic/gin"

type HealthzController interface {
	CheckHealth(ginCtx *gin.Context)
	DbPing(ginCtx *gin.Context)
}

type HealthzService interface {
	DbPing() error
}
