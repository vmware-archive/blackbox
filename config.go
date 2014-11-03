package blackbox

import (
	"os"

	"github.com/fraenkel/candiedyaml"
)

type Drain struct {
	Transport string `yaml:"transport"`
	Address   string `yaml:"address"`
}

type Config struct {
	Destination Drain    `yaml:"destination"`
	Files       []string `yaml:"files"`
}

func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config

	if err := candiedyaml.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
