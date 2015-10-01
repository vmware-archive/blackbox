package integration_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/concourse/blackbox/integration"

	sl "github.com/ziutek/syslog"

	"github.com/concourse/blackbox"
	"github.com/concourse/blackbox/syslog"
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
			Syslog: blackbox.SyslogConfig{
				Destination: syslog.Drain{
					Transport: "udp",
					Address:   syslogServer.Addr,
				},
				Sources: []blackbox.SyslogSource{
					{
						Path: filePathToWatch,
						Tag:  "test-tag",
					},
				},
			},
		}
	}

	buildConfig := func(filePathToWatch string) blackbox.Config {
		return buildConfigHostname("", filePathToWatch)
	}

	It("logs any new lines of a watched file to syslog", func() {
		fileToWatch, err := ioutil.TempFile("", "tail")
		Expect(err).NotTo(HaveOccurred())

		config := buildConfig(fileToWatch.Name())
		blackboxRunner.StartWithConfig(config)

		fileToWatch.WriteString("hello\n")
		fileToWatch.WriteString("world\n")
		fileToWatch.Sync()
		fileToWatch.Close()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		Eventually(inbox.Messages, "2s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("world"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		blackboxRunner.Stop()
		fileToWatch.Close()
		os.Remove(fileToWatch.Name())
	})

	It("can have a custom hostname", func() {
		fileToWatch, err := ioutil.TempFile("", "tail")
		Expect(err).NotTo(HaveOccurred())

		config := buildConfigHostname("fake-hostname", fileToWatch.Name())
		blackboxRunner.StartWithConfig(config)

		fileToWatch.WriteString("hello\n")
		fileToWatch.Sync()
		fileToWatch.Close()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring("fake-hostname"))

		blackboxRunner.Stop()
		os.Remove(fileToWatch.Name())
	})

	It("does not log existing messages", func() {
		fileToWatch, err := ioutil.TempFile("", "tail")
		Expect(err).NotTo(HaveOccurred())

		fileToWatch.WriteString("already present\n")
		fileToWatch.Sync()

		config := buildConfig(fileToWatch.Name())
		blackboxRunner.StartWithConfig(config)

		fileToWatch.WriteString("hello\n")
		fileToWatch.Sync()
		fileToWatch.Close()

		var message *sl.Message
		Eventually(inbox.Messages, "2s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))

		blackboxRunner.Stop()
		os.Remove(fileToWatch.Name())
	})
})
