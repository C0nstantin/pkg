package config

import (
	"fmt"
	"os"
	"path"
)

var defaultConfigPath = "./config/config.yaml"

// MustLoadConfig attempts to load the application configuration into the provided struct.
// It leverages LoadConfig for the actual loading process. If LoadConfig returns an error,
// MustLoadConfig will panic, halting the program execution. This function is useful for
// scenarios where a failure to load the configuration is considered fatal and unrecoverable.
//
// Parameters:
//   - cfg: A pointer to the struct into which the configuration should be loaded. This struct
//     should be prepared to receive the configuration data, typically with struct tags
//     defining the mapping from the configuration file or environment variables.
//
// This function does not return a value. Instead, it panics if an error occurs during the
// configuration loading process.
func MustLoadConfig(cfg interface{}) {
	if err := LoadConfig(cfg); err != nil {
		panic(err)
	}
}

// LoadConfig loads the application configuration into the provided struct.
// It first attempts to find a configuration file by checking the "CONFIG" environment variable.
// If the environment variable is not set, it looks for a default configuration file in the
// "./config/config.yaml" path relative to the current working directory. If neither is found,
// it falls back to loading configuration values from environment variables.
//
// The function supports overriding the main configuration file with a local configuration file.
// This local file should have the same name as the main configuration file but with ".local"
// appended before the file extension. For example, "config.yaml" would be overridden by
// "config.local.yaml" if it exists. The local configuration file is optional and is intended
// for development use or to override configuration without modifying the main configuration file.
//
// Parameters:
//   - cfg: A pointer to the struct into which the configuration should be loaded. This struct
//     should be prepared to receive the configuration data, typically with struct tags
//     defining the mapping from the configuration file or environment variables.
//
// Returns:
//   - An error if loading the configuration fails for any reason, such as if the configuration
//     file cannot be found (and no environment variables are set), if there is an error reading
//     the configuration file, or if there is an error applying the configuration to the provided
//     struct. If the configuration is successfully loaded from either the file(s) or environment
//     variables, nil is returned.
func LoadConfig(cfg interface{}) error {
	// cfg only point to struct
	configFile, exists := os.LookupEnv("CONFIG")
	if !exists {
		configFile = defaultConfigPath
	}
	currentDir, _ := os.Getwd()
	configFile = path.Join(currentDir, configFile)
	if _, err := os.Stat(configFile); err == nil {
		err = readFileWithLocal(configFile, cfg)
		if err != nil {
			return err
		}
	}

	return ReadEnv(cfg)
}

func readFileWithLocal(configFile string, cfg interface{}) error {
	err := ReadConfig(configFile, cfg)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	localConfigFile := configFile[:len(configFile)-len(path.Ext(configFile))] + ".local" + path.Ext(configFile)
	if _, err = os.Stat(localConfigFile); err == nil {
		err := ReadConfig(localConfigFile, cfg)
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
	}

	return nil
}
