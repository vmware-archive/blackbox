package blackbox

import (
	"os"

	"github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/concourse/blackbox/syslog"
)

type Source struct {
	Path string `yaml:"path"`
	Tag  string `yaml:"tag"`
}

type SyslogConfig struct {
	Destination syslog.Drain `yaml:"destination"`
	Sources     []Source     `yaml:"sources"`
}

type Config struct {
	Hostname string `yaml:"hostname"`

	SyslogConfig SyslogConfig `yaml:"syslog"`
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

	if config.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		config.Hostname = hostname
	}

	return &config, nil
}
