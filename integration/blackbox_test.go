package integration_test

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/blackbox"
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

		// The socket seems to stay open and grab adjacent test messages.
		// Sleep-driven development until this is fixed.
		time.Sleep(1 * time.Second)
	})

	It("logs any new lines of a watched file to syslog", func() {
		hostname, err := os.Hostname()
		Ω(err).ShouldNot(HaveOccurred())

		fileToWatch, err := ioutil.TempFile("", "tail")
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
		Ω(message.Content).Should(ContainSubstring(hostname))

		Eventually(inbox.Messages).Should(Receive(&message))
		Ω(message.Content).Should(ContainSubstring("world"))
		Ω(message.Content).Should(ContainSubstring("test-tag"))
		Ω(message.Content).Should(ContainSubstring(hostname))

		blackboxRunner.Stop()
		fileToWatch.Close()
		os.Remove(fileToWatch.Name())
	})

	It("can have a custom hostname", func() {
		fileToWatch, err := ioutil.TempFile("", "tail")
		Ω(err).ShouldNot(HaveOccurred())

		config := blackbox.Config{
			Hostname: "fake-hostname",
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
		Consistently(inbox.Messages).ShouldNot(Receive())

		blackboxRunner.Stop()
		fileToWatch.Close()
		os.Remove(fileToWatch.Name())
	})
})
