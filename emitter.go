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

	for {
		expvars, err := e.expvar.Fetch()
		if err != nil {
			log.Println("failed to fetch expvars: %s", err)
			continue
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
			log.Println("failed publish series: %s", err)
		}

		time.Sleep(e.interval)
	}

	return nil
}
