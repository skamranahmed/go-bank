package cmd

import (
	"github.com/skamranahmed/go-bank/cmd/router"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func Run() error {
	logger.Init()

	services, err := internal.BootstrapServices()
	if err != nil {
		return err
	}

	router := router.Init(services)
	server.Start(router)

	return nil
}
