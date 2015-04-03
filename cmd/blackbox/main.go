package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	"github.com/concourse/blackbox"
	"github.com/concourse/blackbox/datadog"
	"github.com/concourse/blackbox/expvar"
	"github.com/concourse/blackbox/syslog"
)

var configPath = flag.String(
	"config",
	"",
	"path to the configuration file",
)

func main() {
	flag.Parse()

	logger := log.New(os.Stderr, "", log.LstdFlags)

	if *configPath == "" {
		logger.Fatalln("-config must be specified")
	}

	config, err := blackbox.LoadConfig(*configPath)
	if err != nil {
		logger.Fatalf("could not load config file: %s\n", err)
	}

	members := grouper.Members{}

	if len(config.Syslog.Sources) > 0 {
		drainer, err := syslog.NewDrainer(
			config.Syslog.Destination,
			config.Hostname,
		)
		if err != nil {
			logger.Fatalf("could not drain to syslog: %s\n", err)
		}

		members = append(members, buildTailers(config.Syslog.Sources, drainer)...)
	}

	if len(config.Expvar.Sources) > 0 {
		datadogClient := datadog.NewClient(config.Expvar.Datadog.APIKey)

		members = append(members, buildEmitters(config.Hostname, config.Expvar, datadogClient)...)
	}

	group := grouper.NewParallel(nil, members)
	running := ifrit.Invoke(sigmon.New(group))

	err = <-running.Wait()
	if err != nil {
		logger.Fatalf("failed: %s", err)
	}
}

func buildTailers(sources []blackbox.SyslogSource, drainer syslog.Drainer) grouper.Members {
	members := make(grouper.Members, len(sources))

	for i, source := range sources {
		tailer := &blackbox.Tailer{
			Source:  source,
			Drainer: drainer,
		}

		members[i] = grouper.Member{source.Path, tailer}
	}

	return members
}

func buildEmitters(hostname string, config blackbox.ExpvarConfig, datadogClient datadog.Client) grouper.Members {
	members := make(grouper.Members, len(config.Sources))

	for i, source := range config.Sources {
		fetcher := expvar.NewFetcher(source.URL)
		emitter := blackbox.NewEmitter(
			datadogClient,
			fetcher,
			time.Duration(config.Interval),
			hostname,
			source.Tags,
		)

		members[i] = grouper.Member{source.Name, emitter}
	}

	return members
}
