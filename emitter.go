package blackbox

import (
	"log"
	"os"
	"time"

	"github.com/concourse/blackbox/datadog"
	"github.com/concourse/blackbox/expvar"
)

type emitter struct {
	datadog datadog.Client
	expvar  expvar.Fetcher

	interval time.Duration
	host     string
	tags     []string
}

func NewEmitter(
	datadog datadog.Client,
	expvar expvar.Fetcher,
	interval time.Duration,
	host string,
	tags []string,
) *emitter {
	return &emitter{
		datadog:  datadog,
		expvar:   expvar,
		interval: interval,
		host:     host,
		tags:     tags,
	}
}

func (e *emitter) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)

	ticker := time.NewTicker(e.interval)

	for {
		expvars, err := e.expvar.Fetch()
		if err != nil {
			log.Printf("failed to fetch expvars: %s\n", err)

			select {
			case <-ticker.C:
				continue
			case <-signals:
				ticker.Stop()
				return nil
			}
		}

		series := make(datadog.Series, 0, expvars.Size())
		now := time.Now()

		expvars.Walk(func(path string, value float32) {
			series = append(series, datadog.Metric{
				Name: path,
				Points: []datadog.Point{
					{Timestamp: now, Value: value},
				},
				Host: e.host,
				Tags: e.tags,
			})
		})

		if err := e.datadog.PublishSeries(series); err != nil {
			log.Printf("failed to publish series: %s\n", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-signals:
			ticker.Stop()
			return nil
		}
	}
}
