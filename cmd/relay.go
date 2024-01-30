package cmd

import (
	"fmt"

	"github.com/stafiprotocol/neutron-lsd-relay/pkg/config"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/log"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/utils"
	"github.com/stafiprotocol/neutron-lsd-relay/task"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagLogLevel   = "log_level"
	flagConfigPath = "config"

	defaultKeystorePath          = "./keys"
	defaultConfigPath            = "./config.toml"
	defaultLogPath               = "./log_data"
	defaultBackendOptions        = "test"
	defaultTimeTicket     uint32 = 10
	defaultLogLevel              = logrus.InfoLevel.String()
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Args:  cobra.ExactArgs(0),
		Short: "Start lsd relay",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)
			// check log level
			logLevelStr, err := cmd.Flags().GetString(flagLogLevel)
			if err != nil {
				return err
			}
			logLevel, err := logrus.ParseLevel(logLevelStr)
			if err != nil {
				return err
			}
			fmt.Printf("log level: %s\n", logLevelStr)
			logrus.SetLevel(logLevel)

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			if cfg.BackendOptions == "" {
				cfg.BackendOptions = defaultBackendOptions
			}
			if cfg.TaskTicker == 0 {
				cfg.TaskTicker = defaultTimeTicket
			}
			if cfg.KeystorePath == "" {
				cfg.KeystorePath = defaultKeystorePath
			}
			if cfg.LogFilePath == "" {
				cfg.LogFilePath = defaultLogPath
			}

			err = log.InitLogFile(cfg.LogFilePath)
			if err != nil {
				return err
			}
			logrus.Infof("cfg: %+v", cfg)

			ctx := utils.ShutdownListener()

			logrus.Info("task starting...")
			t, err := task.NewTask(cfg)
			if err != nil {
				return err
			}
			err = t.Start()
			if err != nil {
				logrus.Errorf("task start err: %s", err)
				return err
			}
			defer func() {
				logrus.Infof("shutting down task ...")
				t.Stop()
			}()

			<-ctx.Done()
			return nil
		},
	}

	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	cmd.Flags().String(flagLogLevel, defaultLogLevel, "The logging level (trace|debug|info|warn|error|fatal|panic)")

	return cmd
}
