package main

import (
	"context"
	"flag"
	"os"

	"github.com/skamranahmed/go-bank/cmd"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func main() {
	role := flag.String("role", cmd.RoleServer, "role to run: server, worker-default or worker-priority")

	// parse the flags
	flag.Parse()

	err := cmd.Run(*role)
	if err != nil {
		logger.Error(context.TODO(), "Error during server startup: %+v", err)
		os.Exit(1)
	}
}
