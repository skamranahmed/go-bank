package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/skamranahmed/go-bank/cmd/middleware"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	authenticationController "github.com/skamranahmed/go-bank/internal/authentication/controller"
	healthzController "github.com/skamranahmed/go-bank/internal/healthz/controller"
	"github.com/skamranahmed/go-bank/pkg/metrics"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func Init(db *bun.DB, services *internal.Services) *gin.Engine {
	environment := config.GetEnvironment()

	// register prometheus metrics
	metrics.Register()

	if environment == config.APP_ENVIRONMENT_LOCAL {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	/*
		OpenTelemetry middleware is important to be registered
		before the request logger middleware because we need to
		extract the trace from the request context in the request
		logger middleware for attaching the correlation_id to the root span
	*/
	telemetryServiceName := config.GetTelemetryConfig().ServiceName
	router.Use(otelgin.Middleware(telemetryServiceName))

	/*
		The request logger middleware is disabled in the test environment
		because it creates a lot of logs when running tests in verbose mode
		and hence makes it difficult to read the test output.
	*/
	if environment != config.APP_ENVIRONMENT_TEST {
		router.Use(middleware.RequestLoggerMiddleware())
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	healthzController.Register(router, healthzController.Dependency{
		HealthzService: services.HealthzService,
	})

	authenticationController.Register(router, authenticationController.Dependency{
		Db:                    db,
		AuthenticationService: services.AuthenticationService,
		UserService:           services.UserService,
		AccountService:        services.AccountService,
		TaskEnqueuer:          services.TaskEnqueuer,
	})

	return router
}
