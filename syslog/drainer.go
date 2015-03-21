package syslog

import (
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
)

type Drain struct {
	Transport string `yaml:"transport"`
	Address   string `yaml:"address"`
}

//go:generate counterfeiter . Drainer

type Drainer interface {
	Drain(line string, tag string) error
}

type drainer struct {
	logger   *syslog.Logger
	hostname string
}

func NewDrainer(drain Drain, hostname string) (*drainer, error) {
	logger, err := syslog.Dial(
		hostname,
		drain.Transport,
		drain.Address,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &drainer{
		logger:   logger,
		hostname: hostname,
	}, nil
}

func (d *drainer) Drain(line string, tag string) error {
	d.logger.Packets <- syslog.Packet{
		Severity: syslog.SevInfo,
		Facility: syslog.LogUser,
		Hostname: d.hostname,
		Tag:      tag,
		Time:     time.Now(),
		Message:  line,
	}

	select {
	case err := <-d.logger.Errors:
		return err
	default:
		return nil
	}
}
