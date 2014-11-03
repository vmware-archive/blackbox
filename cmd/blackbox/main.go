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

	drainer, err := blackbox.NewDrainer(config.Destination)
	if err != nil {
		logger.Fatalf("could not drain to syslog: %s\n", err)
	}

	members := buildTailers(config.Files, drainer)

	group := grouper.NewParallel(os.Interrupt, members)
	running := ifrit.Invoke(
		sigmon.New(group),
	)

	err = <-running.Wait()

	tail.Cleanup()
	drainer.Close()

	if err != nil {
		logger.Fatalf("failed: %s", err)
	}
}

func buildTailers(filepaths []string, drainer *blackbox.Drainer) grouper.Members {
	members := make(grouper.Members, len(filepaths))

	for i, filepath := range filepaths {
		tailer := &blackbox.Tailer{
			Path:    filepath,
			Drainer: drainer,
		}

		members[i] = grouper.Member{filepath, tailer}
	}

	return members
}
