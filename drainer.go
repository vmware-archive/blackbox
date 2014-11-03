package blackbox

import "log/syslog"

const defaultPriority = syslog.LOG_INFO | syslog.LOG_LOCAL0

type Drainer struct {
	writer *syslog.Writer
}

func NewDrainer(drain Drain) (*Drainer, error) {
	writer, err := syslog.Dial(drain.Transport, drain.Address, defaultPriority, "blackbox")
	if err != nil {
		return nil, err
	}

	return &Drainer{
		writer: writer,
	}, nil
}

func (d *Drainer) Drain(line string) error {
	return d.writer.Info(line)
}

func (d *Drainer) Close() error {
	return d.writer.Close()
}
