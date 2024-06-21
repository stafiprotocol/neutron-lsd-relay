package cmd

import (
	"fmt"
	"os"

	"github.com/stafiprotocol/neutron-lsd-relay/pkg/config"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/log"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/utils"
	"github.com/stafiprotocol/neutron-lsd-relay/task"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagLogLevel = "log-level"
	flagBasePath = "base-path"

	defaultBasePath = "~/cosmos-stack"
	defaultLogLevel = logrus.InfoLevel.String()
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Args:  cobra.ExactArgs(0),
		Short: "Start lsd relay",
		RunE: func(cmd *cobra.Command, args []string) error {
			basePath, err := cmd.Flags().GetString(flagBasePath)
			if err != nil {
				return err
			}
			basePath, err = utils.ReplaceUserHomeDir("--base-path", basePath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(basePath)
			if err != nil {
				return err
			}

			cfg.KeyringDir, err = utils.ReplaceUserHomeDir("keyringDir", cfg.KeyringDir)
			if err != nil {
				return err
			}

			keyPath := cfg.KeyringDir + "/keyring-file/" + cfg.KeyName + ".info"
			if _, err := os.Stat(keyPath); err != nil {
				return fmt.Errorf("please import your account first, key path: %s", keyPath)
			}

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

			if cfg.KeyringBackend == "" {
				return fmt.Errorf("backend_options must be set")
			}
			if cfg.TaskTicker == 0 {
				return fmt.Errorf("task_ticker must be set")
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

	cmd.Flags().String(flagLogLevel, defaultLogLevel, "The logging level (trace|debug|info|warn|error|fatal|panic)")
	cmd.Flags().String(flagBasePath, defaultBasePath, "base path a directory where your config.toml resids")

	return cmd
}
