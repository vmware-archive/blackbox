package integration_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"github.com/concourse/blackbox"
	"github.com/fraenkel/candiedyaml"
	"github.com/ziutek/syslog"
)

var blackboxPath string

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blackbox Suite")
}

var _ = BeforeSuite(func() {
	var err error
	blackboxPath, err = gexec.Build("github.com/concourse/blackbox/cmd/blackbox")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

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

	err = candiedyaml.NewEncoder(configFile).Encode(config)
	Ω(err).ShouldNot(HaveOccurred())

	return configFile.Name()
}
