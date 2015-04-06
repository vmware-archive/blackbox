package blackbox

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/concourse/blackbox/syslog"
	"gopkg.in/yaml.v2"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var num int64
	if err := unmarshal(&num); err == nil {
		*d = Duration(num)
		return nil
	}

	var str string
	if err := unmarshal(&str); err != nil {
		return errors.New("invalid duration; must be string or number")
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*d = Duration(duration)

	return nil
}

type SyslogSource struct {
	Path string `yaml:"path"`
	Tag  string `yaml:"tag"`
}

type SyslogConfig struct {
	Destination syslog.Drain   `yaml:"destination"`
	Sources     []SyslogSource `yaml:"sources"`
}

type ExpvarSource struct {
	Name string   `yaml:"name"`
	URL  string   `yaml:"url"`
	Tags []string `yaml:"tags"`
}

type DatadogConfig struct {
	APIKey string `yaml:"api_key"`
}

type ExpvarConfig struct {
	Interval Duration       `yaml:"interval"`
	Datadog  DatadogConfig  `yaml:"datadog"`
	Sources  []ExpvarSource `yaml:"sources"`
}

type Config struct {
	Hostname string `yaml:"hostname"`

	Syslog SyslogConfig `yaml:"syslog"`
	Expvar ExpvarConfig `yaml:"expvar"`
}

func LoadConfig(path string) (*Config, error) {
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := yaml.Unmarshal(configFile, &config); err != nil {
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
