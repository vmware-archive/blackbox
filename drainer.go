package blackbox

//go:generate counterfeiter . Drainer

import (
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
)

type Drainer interface {
	Drain(line string, tag string) error
}

type SyslogDrainer struct {
	logger   *syslog.Logger
	hostname string
}

func NewSyslogDrainer(drain SyslogDrain, hostname string) (*SyslogDrainer, error) {
	logger, err := syslog.Dial(
		hostname,
		drain.Transport,
		drain.Address,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &SyslogDrainer{
		logger:   logger,
		hostname: hostname,
	}, nil
}

func (d *SyslogDrainer) Drain(line string, tag string) error {
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
