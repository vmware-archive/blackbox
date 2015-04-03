package blackbox

import (
	"log"
	"os"
	"time"

	"github.com/ActiveState/tail"
	"github.com/ActiveState/tail/watch"

	"github.com/concourse/blackbox/syslog"
)

type Tailer struct {
	Source  SyslogSource
	Drainer syslog.Drainer
}

func (tailer *Tailer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	watch.POLL_DURATION = 1 * time.Second

	t, err := tail.TailFile(tailer.Source.Path, tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: os.SEEK_END,
		},
	})

	if err != nil {
		return err
	}
	defer tail.Cleanup()

	close(ready)

	for {
		select {
		case line, ok := <-t.Lines:
			if !ok {
				log.Println("lines flushed; exiting tailer")
				return nil
			}

			tailer.Drainer.Drain(line.Text, tailer.Source.Tag)
		case <-signals:
			return t.Stop()
		}
	}

	return nil
}
