package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"club.asynclab/asrp/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/program"
	"club.asynclab/asrp/program/client"
	"club.asynclab/asrp/program/server"
	"github.com/BurntSushi/toml"
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
				logger.Error("Usage: server <config_file>")
				return
			}

			config := &config.ConfigServer{}
			if _, err := toml.DecodeFile(args[0], config); err != nil {
				logger.Error("Error decoding config file: ", err)
				return
			}
			server := server.NewServer(ctx, config)
			program.Program = server
			server.Run()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "client",
		Short: "Start client",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				logger.Error("Invalid arguments")
				logger.Error("Usage: client <config_file>")
				return
			}

			config := &config.ConfigClient{}
			if _, err := toml.DecodeFile(args[0], config); err != nil {
				logger.Error("Error decoding config file: ", err)
				return
			}
			client := client.NewClient(ctx, config)
			program.Program = client
			client.Run()
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
