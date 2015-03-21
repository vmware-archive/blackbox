package blackbox

import (
	"log"
	"os"
	"time"

	"github.com/concourse/blackbox/datadog"
	"github.com/concourse/blackbox/expvar"
)

type Emitter struct {
	datadog datadog.Client
	expvar  expvar.Fetcher
}

func (e *Emitter) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
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
				Host: "a-host",                 // TODO
				Tags: []string{"cool", "tags"}, // TODO
			})
		})

		if err := e.datadog.PublishSeries(series); err != nil {
			log.Println("failed publish series: %s", err)
		}

		time.Sleep(10 * time.Second) // TODO
	}

	return nil
}
