package router

import (
	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/internal"
	authenticationController "github.com/skamranahmed/go-bank/internal/authentication/controller"
	healthzController "github.com/skamranahmed/go-bank/internal/healthz/controller"
	"github.com/uptrace/bun"
)

func Init(db *bun.DB, services *internal.Services) *gin.Engine {
	router := gin.Default()
	router.Use(gin.Recovery())

	healthzController.Register(router, services.HealthzService)

	authenticationController.Register(router, authenticationController.Dependency{
		Db:                    db,
		AuthenticationService: services.AuthenticationService,
		UserService:           services.UserService,
	})

	return router
}
