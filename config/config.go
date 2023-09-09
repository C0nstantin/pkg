package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

var ErrFileNotExists = errors.New("config file not found, set env variable CONFIG to path config file")
var defaultConfigPath = "./config/config.yaml"

func MustLoadConfig(cfg interface{}) {
	if err := LoadConfig(cfg); err != nil {
		panic(err)
	}
}
func LoadConfig(cfg interface{}) error {
	// cfg only point to struct
	configFile, exists := os.LookupEnv("CONFIG")
	if !exists {
		currentDir, _ := os.Getwd()
		defaultConfig := path.Join(currentDir, defaultConfigPath)
		_, err := os.Stat(defaultConfig)
		switch {
		case err == nil:
			configFile = defaultConfig
		case !errors.Is(err, os.ErrNotExist):
			return fmt.Errorf("undefined error with config file:  %w ", err)
		default:
			log.Println("Warning: config file not found")
			err := cleanenv.ReadEnv(cfg)
			if err != nil {
				return err
			}
			return nil // ErrFileNotExists
		}
	}
	err := cleanenv.ReadConfig(configFile, cfg)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	localConfigFile := configFile[:len(configFile)-len(path.Ext(configFile))] + ".local" + path.Ext(configFile)
	_, err = os.Stat(localConfigFile)
	if err == nil {
		err := cleanenv.ReadConfig(localConfigFile, cfg)
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return err
	}

	return nil
}
