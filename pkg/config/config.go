// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"fmt"
	"os"
	"strings"

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

func Load(basePath string) (*Config, error) {
	basePath = strings.TrimSuffix(basePath, "/")
	configFilePath := basePath + "/config.toml"
	fmt.Printf("config path: %s\n", configFilePath)

	cfg := Config{}
	if err := loadSysConfig(configFilePath, &cfg); err != nil {
		return nil, err
	}
	cfg.LogFilePath = basePath + "/log_data"
	cfg.KeystorePath = KeyStoreFilePath(basePath)

	return &cfg, nil
}

func KeyStoreFilePath(basePath string) string {
	basePath = strings.TrimSuffix(basePath, "/")
	return basePath + "/keystore"
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
