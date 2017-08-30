package main

import (
	"flag"
	"log"
	"os"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	"github.com/concourse/blackbox"
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

	group := grouper.NewDynamic(nil, 0, 0)
	running := ifrit.Invoke(sigmon.New(group))

	if config.Syslog.SourceDir != "" {
		go func() {
			drainerFactory := syslog.NewDrainerFactory(config.Syslog.Destination, config.Hostname)
			fileWatcher := blackbox.NewFileWatcher(logger, config.Syslog.SourceDir, group.Client(), drainerFactory)
			fileWatcher.Watch()
		}()
	}

	for _, dir := range config.Syslog.SourceDirs {
		go func(dir string) {
			drainerFactory := syslog.NewDrainerFactory(config.Syslog.Destination, config.Hostname)
			fileWatcher := blackbox.NewFileWatcher(logger, dir, group.Client(), drainerFactory)
			fileWatcher.Watch()
		}(dir)
	}

	err = <-running.Wait()
	if err != nil {
		logger.Fatalf("failed: %s", err)
	}
}
