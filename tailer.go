package blackbox

import (
	"os"

	"github.com/ActiveState/tail"
)

type Tailer struct {
	Source  Source
	Drainer *SyslogDrainer
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

	close(ready)

	for {
		select {
		case line := <-t.Lines:
			tailer.Drainer.Drain(line.Text, tailer.Source.Tag)
		case <-signals:
			return t.Stop()
		}
	}

	return nil
}
