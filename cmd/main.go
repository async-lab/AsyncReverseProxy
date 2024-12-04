package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/program"
	"club.asynclab/asrp/pkg/program/client"
	"club.asynclab/asrp/pkg/program/server"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var logger = logging.GetLogger()

var ctx, cancel = context.WithCancel(context.Background())

var rootCmd = &cobra.Command{
	Use:   "asrp",
	Short: "Highly available and scalable reverse proxy",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&config.IsVerbose, "verbose", "v", false, "verbose output")

	start := func(path string, cfg config.IConfig, prog program.IProgram) {
		logging.Init(config.IsVerbose)

		if _, err := toml.DecodeFile(path, cfg); err != nil {
			logger.Error("Error decoding config file: ", err)
			return
		}

		program.Program = prog
		prog.Run()
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:     "server config_file",
		Aliases: []string{"s"},
		Short:   "Start server",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := &config.ConfigServer{}
			start(args[0], cfg, server.NewServer(ctx, cfg))
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:     "client config_file",
		Aliases: []string{"c"},
		Short:   "Start client",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := &config.ConfigClient{}
			start(args[0], cfg, client.NewClient(ctx, cfg))
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
