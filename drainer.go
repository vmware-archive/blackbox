package blackbox

import (
	"time"

	"github.com/papertrail/remote_syslog2/syslog"
)

type Drainer struct {
	logger   *syslog.Logger
	hostname string
}

func NewDrainer(drain Drain, hostname string) (*Drainer, error) {
	logger, err := syslog.Dial(
		hostname,
		drain.Transport,
		drain.Address,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &Drainer{
		logger:   logger,
		hostname: hostname,
	}, nil
}

func (d *Drainer) Drain(line string, tag string) error {
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
