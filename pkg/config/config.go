// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	TaskTicker          uint32
	PoolAddr            string
	StakeManager        string
	GasPrice            string
	KeyName             string
	BackendOptions      string
	RunForEntrustedPool bool

	EndpointList []string

	LogFilePath  string
	KeystorePath string
}

func Load(configFilePath string) (*Config, error) {
	cfg := Config{}
	if err := loadSysConfig(configFilePath, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.LogFilePath) == 0 {
		cfg.LogFilePath = "./log_data"
	}

	return &cfg, nil
}

func loadSysConfig(path string, config *Config) error {
	_, err := os.Open(path)
	if err != nil {
		return err
	}
	if _, err := toml.DecodeFile(path, config); err != nil {
		return err
	}
	fmt.Println("load config success")
	return nil
}
