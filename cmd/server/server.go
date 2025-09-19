package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func Start(router *gin.Engine) {
	serverConfig := config.GetServerConfig()
	server := newServer(
		fmt.Sprintf(":%d", serverConfig.Port),
		router,
	)

	// start server in background
	go server.start()

	// block until SIGTERM or SIGINT signal is received
	server.waitForSignal()

	// this context is used to inform the server that it has `X` seconds to process the request(s) it is currently handling
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(time.Duration(serverConfig.GracefulShutdownTimeoutInSeconds)*time.Second),
	)
	defer cancel()

	err := server.stop(ctx)
	if err != nil {
		logger.Fatal(ctx, "Error while server shutdown, doing it forcefully now")
	}

	logger.Info(ctx, "Server is stopping")
}

type apiServer struct {
	*http.Server
}

func newServer(address string, handler http.Handler) *apiServer {
	return &apiServer{
		Server: &http.Server{
			Addr:    address,
			Handler: handler,
		},
	}
}

func (s *apiServer) start() {
	ctx := context.TODO()

	logger.Info(ctx, "Server listening on port %v", config.GetServerConfig().Port)

	err := s.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Error(ctx, "Error while starting server: %v", err)
		logger.Fatal(ctx, "Server is stopping")
	}
}

func (s *apiServer) waitForSignal() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	signalValue := <-done
	logger.Info(context.TODO(), "Received '%+v' syscall, gracefully shutting down server", signalValue.String())
}

func (s *apiServer) stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
