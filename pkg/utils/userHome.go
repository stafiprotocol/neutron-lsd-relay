package utils

import (
	"fmt"
	"os"
	"strings"
)

func ReplaceUserHomeDir(configName, path string) (string, error) {
	userHomeDir := os.Getenv("HOME")
	if strings.HasPrefix(path, "~") {
		if userHomeDir == "" {
			return "", fmt.Errorf("please use absolute path for %s or config HOME environment", configName)
		}
		return strings.Replace(path, "~", userHomeDir, 1), nil
	}
	return path, nil
}
