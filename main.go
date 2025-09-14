package main

import (
	"os"

	"github.com/skamranahmed/go-bank/cmd"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func main() {
	err := cmd.Run()
	if err != nil {
		logger.Error("Error during server startup: %+v", err)
		os.Exit(1)
	}
}
