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

func loadConfig[T config.IConfig](path string) (T, error) {
	cfg := new(T)
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return *cfg, err
	}
	return *cfg, nil
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&config.IsVerbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(&cobra.Command{
		Use:     "server config_file",
		Aliases: []string{"s"},
		Short:   "Start server",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logging.Init(config.IsVerbose)
			cfg, err := loadConfig[config.ConfigServer](args[0])
			if err != nil {
				logger.Error("Error loading config: ", err)
				return
			}
			s, err := server.NewServer(ctx, cfg)
			if err != nil {
				logger.Error("Error creating server: ", err)
				return
			}
			program.Program = s
			s.Run()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:     "client config_file",
		Aliases: []string{"c"},
		Short:   "Start client",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logging.Init(config.IsVerbose)
			cfg, err := loadConfig[config.ConfigClient](args[0])
			if err != nil {
				logger.Error("Error loading config: ", err)
				return
			}
			c := client.NewClient(ctx, cfg)
			program.Program = c
			c.Run()
		},
	})

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-interruptChan
		logger.Info("Interrupt signal received.")
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
