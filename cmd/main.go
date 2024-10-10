package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/program"
)

var logger = logging.GetLogger()

var ctx, cancel = context.WithCancel(context.Background())

func init() {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-interruptChan
		logger.Info("Interrupt signal received")
		cancel()
	}()
}

func main() {
	args := os.Args

	if len(args) > 2 {
		switch args[1] {
		case "client":
			client := program.NewClient(ctx, args[2])
			client.Hello()
			client.StartProxy(":54321", "localhost:80")
		case "server":
			server := program.NewServer(ctx, args[2])
			server.Listen()
		default:
			logger.Error("Invalid argument")
			return
		}
	} else {
		logger.Error("Invalid argument for number")
	}

	logger.Info("Exiting...")
	os.Exit(0)
}
