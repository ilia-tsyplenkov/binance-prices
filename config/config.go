package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Address      string   `yaml:"address"`
	BinanceWsURL string   `yaml:"binance_ws_url"`
	Tickers      []string `yaml:"tickers"`
}

func parseFlags() (string, error) {
	var configPath string

	flag.StringVar(&configPath, "config", "./config/config.yaml", "path to config file")

	flag.Parse()

	if err := validateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}

func Init() (*Config, error) {
	cfgPath, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := newConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	return cfg, err
}

func newConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
