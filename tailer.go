package blackbox

import (
	"os"

	"github.com/ActiveState/tail"

	"github.com/concourse/blackbox/syslog"
)

type Tailer struct {
	Source  SyslogSource
	Drainer syslog.Drainer
}

func (tailer *Tailer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	t, err := tail.TailFile(tailer.Source.Path, tail.Config{
		Follow: true,
		ReOpen: true,
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
				println("lines flushed; exiting tailer")
				return nil
			}

			tailer.Drainer.Drain(line.Text, tailer.Source.Tag)
		case <-signals:
			return t.Stop()
		}
	}

	return nil
}
