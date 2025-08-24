package router

import (
	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/internal/authentication"
	"github.com/skamranahmed/go-bank/internal/healthz"
	"github.com/uptrace/bun"
)

func Init(db *bun.DB, services *internal.Services) *gin.Engine {
	router := gin.Default()
	router.Use(gin.Recovery())

	healthz.Register(router, services.HealthzService)

	authentication.Register(router, authentication.Dependency{
		Db:                    db,
		AuthenticationService: services.AuthenticationService,
		UserService:           services.UserService,
	})

	return router
}
