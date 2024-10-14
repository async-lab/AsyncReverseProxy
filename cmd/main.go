package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/program"
	"club.asynclab/asrp/program/client"
	"club.asynclab/asrp/program/server"
	"github.com/spf13/cobra"
)

var logger = logging.GetLogger()

var ctx, cancel = context.WithCancel(context.Background())

var rootCmd = &cobra.Command{
	Use:   "asrp",
	Short: "hi!",
	Long:  `HI!`,
}

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "Start server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				logger.Error("Invalid arguments")
				logger.Error("Usage: server <listen_address>")
				return
			}
			server := server.NewServer(ctx, args[0])
			program.Program = server
			server.Listen()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "client",
		Short: "Start client",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 4 {
				logger.Error("Invalid arguments")
				logger.Error("Usage: client <remote_address> <name> <frontend_address> <backend_address>")
				return
			}
			client := client.NewClient(ctx, args[0])
			program.Program = client
			client.Hello()
			client.StartProxy(args[1], args[2], args[3])
		},
	})

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-interruptChan
		logger.Info("Interrupt signal received")
		cancel()
	}()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}

	logger.Info("Exiting...")
	os.Exit(0)
}
