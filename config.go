package blackbox

import (
	"os"

	"github.com/fraenkel/candiedyaml"
)

type Drain struct {
	Transport string `yaml:"transport"`
	Address   string `yaml:"address"`
}

type Source struct {
	Path string `yaml:"path"`
	Tag  string `yaml:"tag"`
}

type Config struct {
	Destination Drain    `yaml:"destination"`
	Sources     []Source `yaml:"sources"`
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