package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const AppName = "postiz-cli"

var ErrConfigDir = errors.New("config directory error")

func ConfigDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("%w: get home directory: %w", ErrConfigDir, err)
		}

		baseDir = filepath.Join(homeDir, "Library", "Application Support", AppName)
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%w: APPDATA not set", ErrConfigDir)
		}

		baseDir = filepath.Join(appData, AppName)
	default:
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("%w: get home directory: %w", ErrConfigDir, err)
			}

			configHome = filepath.Join(homeDir, ".config")
		}

		baseDir = filepath.Join(configHome, AppName)
	}

	return baseDir, nil
}

func EnsureConfigDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return "", fmt.Errorf("%w: create config directory: %w", ErrConfigDir, err)
	}

	return configDir, nil
}

func EnsureKeyringDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	keyringDir := filepath.Join(configDir, "keyring")

	if err := os.MkdirAll(keyringDir, 0o700); err != nil {
		return "", fmt.Errorf("%w: create keyring directory: %w", ErrConfigDir, err)
	}

	return keyringDir, nil
}

func NormalizeEnvVarName(cliName string) string {
	return strings.ToUpper(strings.ReplaceAll(cliName, "-", "_"))
}
