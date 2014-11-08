package main

import (
	"flag"
	"log"
	"os"

	"github.com/ActiveState/tail"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	"github.com/concourse/blackbox"
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

	hostname, _ := os.Hostname()
	drainer, err := blackbox.NewDrainer(config.Destination, hostname)
	if err != nil {
		logger.Fatalf("could not drain to syslog: %s\n", err)
	}

	members := buildTailers(config.Sources, drainer)

	group := grouper.NewParallel(os.Interrupt, members)
	running := ifrit.Invoke(
		sigmon.New(group),
	)
	defer tail.Cleanup()

	err = <-running.Wait()

	if err != nil {
		logger.Fatalf("failed: %s", err)
	}
}

func buildTailers(sources []blackbox.Source, drainer *blackbox.Drainer) grouper.Members {
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
