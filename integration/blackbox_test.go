package integration_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/concourse/blackbox/integration"

	"github.com/concourse/blackbox"
	"github.com/ziutek/syslog"
)

var _ = Describe("Blackbox", func() {
	var blackboxRunner *BlackboxRunner
	var syslogServer *SyslogServer
	var inbox *Inbox

	BeforeEach(func() {
		inbox = NewInbox()
		syslogServer = NewSyslogServer(inbox)
		syslogServer.Start()

		blackboxRunner = NewBlackboxRunner(blackboxPath)
	})

	AfterEach(func() {
		syslogServer.Stop()
	})

	buildConfigHostname := func(hostname string, filePathToWatch string) blackbox.Config {
		return blackbox.Config{
			Hostname: hostname,
			Destination: blackbox.Drain{
				Transport: "udp",
				Address:   syslogServer.Addr,
			},
			Sources: []blackbox.Source{
				{
					Path: filePathToWatch,
					Tag:  "test-tag",
				},
			},
		}
	}

	buildConfig := func(filePathToWatch string) blackbox.Config {
		return buildConfigHostname("", filePathToWatch)
	}

	It("logs any new lines of a watched file to syslog", func() {
		fileToWatch, err := ioutil.TempFile("", "tail")
		Ω(err).ShouldNot(HaveOccurred())

		config := buildConfig(fileToWatch.Name())
		blackboxRunner.StartWithConfig(config)

		fileToWatch.WriteString("hello\n")
		fileToWatch.WriteString("world\n")

		var message *syslog.Message
		Eventually(inbox.Messages).Should(Receive(&message))
		Ω(message.Content).Should(ContainSubstring("hello"))
		Ω(message.Content).Should(ContainSubstring("test-tag"))
		Ω(message.Content).Should(ContainSubstring(Hostname()))

		Eventually(inbox.Messages).Should(Receive(&message))
		Ω(message.Content).Should(ContainSubstring("world"))
		Ω(message.Content).Should(ContainSubstring("test-tag"))
		Ω(message.Content).Should(ContainSubstring(Hostname()))

		blackboxRunner.Stop()
		fileToWatch.Close()
		os.Remove(fileToWatch.Name())
	})

	It("can have a custom hostname", func() {
		fileToWatch, err := ioutil.TempFile("", "tail")
		Ω(err).ShouldNot(HaveOccurred())

		config := buildConfigHostname("fake-hostname", fileToWatch.Name())
		blackboxRunner.StartWithConfig(config)

		fileToWatch.WriteString("hello\n")

		var message *syslog.Message
		Eventually(inbox.Messages).Should(Receive(&message))
		Ω(message.Content).Should(ContainSubstring("hello"))
		Ω(message.Content).Should(ContainSubstring("test-tag"))
		Ω(message.Content).Should(ContainSubstring("fake-hostname"))

		blackboxRunner.Stop()
		fileToWatch.Close()
		os.Remove(fileToWatch.Name())
	})

	It("does not log existing messages", func() {
		fileToWatch, err := ioutil.TempFile("", "tail")
		Ω(err).ShouldNot(HaveOccurred())

		fileToWatch.WriteString("already present\n")

		config := buildConfig(fileToWatch.Name())
		blackboxRunner.StartWithConfig(config)

		fileToWatch.WriteString("hello\n")

		var message *syslog.Message
		Eventually(inbox.Messages).Should(Receive(&message))
		Ω(message.Content).Should(ContainSubstring("hello"))
		Ω(message.Content).Should(ContainSubstring("test-tag"))

		blackboxRunner.Stop()
		fileToWatch.Close()
		os.Remove(fileToWatch.Name())
	})
})
