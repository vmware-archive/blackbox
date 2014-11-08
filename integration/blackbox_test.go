package integration_test

import (
	"io/ioutil"
	"os"

	"github.com/concourse/blackbox"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ziutek/syslog"
)

var _ = Describe("Blackbox", func() {
	var blackboxRunner *BlackboxRunner
	var syslogServer *syslog.Server
	var inbox *Inbox

	BeforeEach(func() {
		inbox = NewInbox()

		syslogServer = syslog.NewServer()
		syslogServer.AddHandler(inbox)
		syslogServer.Listen("127.0.0.1:8742")

		blackboxRunner = NewBlackboxRunner(blackboxPath)
	})

	AfterEach(func() {
		syslogServer.Shutdown()
	})

	It("logs any new lines of a watched file to syslog", func() {
		fileToWatch, err := ioutil.TempFile("", "blackbox")
		Ω(err).ShouldNot(HaveOccurred())

		config := blackbox.Config{
			Destination: blackbox.Drain{
				Transport: "udp",
				Address:   "127.0.0.1:8742",
			},
			Sources: []blackbox.Source{
				{
					Path: fileToWatch.Name(),
					Tag:  "test-tag",
				},
			},
		}

		blackboxRunner.StartWithConfig(config)

		fileToWatch.WriteString("hello\n")
		fileToWatch.WriteString("world\n")

		var message *syslog.Message
		Eventually(inbox.Messages).Should(Receive(&message))
		Ω(message.Content).Should(ContainSubstring("hello"))
		Ω(message.Content).Should(ContainSubstring("test-tag"))

		Eventually(inbox.Messages).Should(Receive(&message))
		Ω(message.Content).Should(ContainSubstring("world"))
		Ω(message.Content).Should(ContainSubstring("test-tag"))

		blackboxRunner.Stop()
		fileToWatch.Close()
		os.Remove(fileToWatch.Name())
	})
})
