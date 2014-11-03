package blackbox

import "log/syslog"

const defaultPriority = syslog.LOG_INFO | syslog.LOG_LOCAL0

type writerPool struct {
	drain Drain
	pool  map[string]*syslog.Writer
}

func (pool *writerPool) GetOrBuild(tag string) (*syslog.Writer, error) {
	var err error
	writer, found := pool.pool[tag]

	if !found {
		writer, err = syslog.Dial(pool.drain.Transport, pool.drain.Address, defaultPriority, tag)
		if err != nil {
			return nil, err
		}

		pool.pool[tag] = writer
	}

	return writer, nil
}

func (pool *writerPool) Close() error {
	for _, writer := range pool.pool {
		writer.Close()
	}

	return nil
}

type Drainer struct {
	pool *writerPool
}

func NewDrainer(drain Drain) (*Drainer, error) {

	return &Drainer{
		pool: &writerPool{
			drain: drain,
			pool:  make(map[string]*syslog.Writer),
		},
	}, nil
}

func (d *Drainer) Drain(line string) error {
	writer, err := d.writer("blackbox")
	if err != nil {
		return err
	}

	return writer.Info(line)
}

func (d *Drainer) Close() error {
	return d.pool.Close()
}

func (d *Drainer) writer(tag string) (*syslog.Writer, error) {
	return d.pool.GetOrBuild(tag)
}
