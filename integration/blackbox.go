package integration

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"

	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/ziutek/syslog"

	"github.com/concourse/blackbox"
)

type SyslogServer struct {
	Addr string

	server *syslog.Server
}

func NewSyslogServer(inbox *Inbox) *SyslogServer {
	server := syslog.NewServer()
	server.AddHandler(inbox)

	return &SyslogServer{
		server: server,
	}
}

func (s *SyslogServer) Start() {
	l, err := net.Listen("tcp", ":0")
	Ω(err).ShouldNot(HaveOccurred())
	l.Close()

	_, port, err := net.SplitHostPort(l.Addr().String())
	Ω(err).ShouldNot(HaveOccurred())

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	s.server.Listen(addr)

	s.Addr = addr
}

func (s *SyslogServer) Stop() {
	s.server.Shutdown()
	s.Addr = ""
}

type Inbox struct {
	Messages chan *syslog.Message
}

func NewInbox() *Inbox {
	return &Inbox{
		Messages: make(chan *syslog.Message),
	}
}

func (i *Inbox) Handle(m *syslog.Message) *syslog.Message {
	if m == nil {
		close(i.Messages)
		return nil
	}

	i.Messages <- m
	return nil
}

type BlackboxRunner struct {
	blackboxPath    string
	blackboxProcess ifrit.Process
}

func NewBlackboxRunner(blackboxPath string) *BlackboxRunner {
	return &BlackboxRunner{
		blackboxPath: blackboxPath,
	}
}

func (runner *BlackboxRunner) StartWithConfig(config blackbox.Config) {
	configPath := runner.createConfigFile(config)

	blackboxCmd := exec.Command(runner.blackboxPath, "-config", configPath)
	blackboxRunner := ginkgomon.New(
		ginkgomon.Config{
			Name:          "blackbox",
			Command:       blackboxCmd,
			AnsiColorCode: "90m",
			StartCheck:    "Seeked",
			Cleanup: func() {
				os.Remove(configPath)
			},
		},
	)

	runner.blackboxProcess = ginkgomon.Invoke(blackboxRunner)
}

func (runner *BlackboxRunner) Stop() {
	ginkgomon.Interrupt(runner.blackboxProcess)
}

func (runner *BlackboxRunner) createConfigFile(config blackbox.Config) string {
	configFile, err := ioutil.TempFile("", "blackbox_config")
	Ω(err).ShouldNot(HaveOccurred())
	defer configFile.Close()

	yamlToWrite, err := yaml.Marshal(config)
	Ω(err).ShouldNot(HaveOccurred())

	configFile.Write(yamlToWrite)

	return configFile.Name()
}
