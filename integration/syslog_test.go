package integration_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/concourse/blackbox/integration"

	sl "github.com/ziutek/syslog"

	"github.com/concourse/blackbox"
	"github.com/concourse/blackbox/syslog"
)

var _ = Describe("Blackbox", func() {
	var (
		blackboxRunner *BlackboxRunner
		syslogServer   *SyslogServer
		inbox          *Inbox
		logDir         string
		tagName        string
		logFile        *os.File
	)

	BeforeEach(func() {
		inbox = NewInbox()
		syslogServer = NewSyslogServer(inbox)
		syslogServer.Start()

		blackboxRunner = NewBlackboxRunner(blackboxPath)

		var err error
		logDir, err = ioutil.TempDir("", "syslog-test")
		Expect(err).NotTo(HaveOccurred())

		tagName = "test-tag"
		err = os.Mkdir(filepath.Join(logDir, tagName), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		logFile, err = os.OpenFile(
			filepath.Join(logDir, tagName, "tail.log"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		logFile.Close()

		syslogServer.Stop()
		err := os.RemoveAll(logDir)
		Expect(err).NotTo(HaveOccurred())
	})

	buildConfigHostname := func(hostname string, dirToWatch string) blackbox.Config {
		return blackbox.Config{
			Hostname: hostname,
			Syslog: blackbox.SyslogConfig{
				Destination: syslog.Drain{
					Transport: "udp",
					Address:   syslogServer.Addr,
				},
				SourceDir: dirToWatch,
			},
		}
	}

	buildConfig := func(dirToWatch string) blackbox.Config {
		return buildConfigHostname("", dirToWatch)
	}

	It("logs any new lines of a file in source directory to syslog with subdirectory name as tag", func() {
		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.WriteString("world\n")
		logFile.Sync()
		logFile.Close()

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
	})

	It("can have a custom hostname", func() {
		config := buildConfigHostname("fake-hostname", logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring("fake-hostname"))

		blackboxRunner.Stop()
	})

	It("does not log existing messages", func() {
		logFile.WriteString("already present\n")
		logFile.Sync()

		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()

		var message *sl.Message
		Eventually(inbox.Messages, "2s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))

		blackboxRunner.Stop()
	})

	It("tracks logs in multiple files in source directory", func() {
		anotherLogFile, err := os.OpenFile(
			filepath.Join(logDir, tagName, "another-tail.log"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
		defer anotherLogFile.Close()

		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		anotherLogFile.WriteString("hello from the other side\n")
		anotherLogFile.Sync()

		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello from the other side"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))
	})

	It("skips files not ending in .log", func() {
		anotherLogFile, err := os.OpenFile(
			filepath.Join(logDir, tagName, "another-tail.log"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
		defer anotherLogFile.Close()

		notALogFile, err := os.OpenFile(
			filepath.Join(logDir, tagName, "not-a-log-file.log.1"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
		defer notALogFile.Close()

		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()

		notALogFile.WriteString("john cena\n")
		notALogFile.Sync()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		Consistently(inbox.Messages).ShouldNot(Receive())

		anotherLogFile.WriteString("hello from the other side\n")
		anotherLogFile.Sync()

		notALogFile.WriteString("my time is now\n")
		notALogFile.Sync()

		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello from the other side"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		Consistently(inbox.Messages).ShouldNot(Receive())
	})

	It("tracks files in multiple directories using multiple tags", func() {
		tagName2 := "2-test-2-tag"
		err := os.Mkdir(filepath.Join(logDir, tagName2), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		anotherLogFile, err := os.OpenFile(
			filepath.Join(logDir, tagName2, "another-tail.log"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
		defer anotherLogFile.Close()

		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		anotherLogFile.WriteString("hello from the other side\n")
		anotherLogFile.Sync()

		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello from the other side"))
		Expect(message.Content).To(ContainSubstring("2-test-2-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))
	})

	It("starts tracking logs in newly created files", func() {
		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		anotherLogFile, err := os.OpenFile(
			filepath.Join(logDir, tagName, "another-tail.log"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
		defer anotherLogFile.Close()

		// wait for tailer to pick up file, twice the interval
		time.Sleep(10 * time.Second)

		anotherLogFile.WriteString("hello from the other side\n")
		anotherLogFile.Sync()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello from the other side"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		By("keeping track of old files")
		logFile.WriteString("hello\n")
		logFile.Sync()

		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))
	})

	It("starts tracking logs in newly created files", func() {
		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		err := os.Remove(filepath.Join(logDir, tagName, "tail.log"))
		Expect(err).NotTo(HaveOccurred())

		// wait for tail process to die, tailer interval is 1 sec
		time.Sleep(2 * time.Second)

		anotherLogFile, err := os.OpenFile(
			filepath.Join(logDir, tagName, "tail.log"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
		defer anotherLogFile.Close()

		// wait for tailer to pick up file, twice the interval
		time.Sleep(10 * time.Second)

		anotherLogFile.WriteString("bye\n")
		anotherLogFile.Sync()

		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("bye"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))
	})

	It("ignores subdirectories in tag directories", func() {
		err := os.Mkdir(filepath.Join(logDir, tagName, "ignore-me"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(
			filepath.Join(logDir, tagName, "ignore-me", "and-my-son.log"),
			[]byte("some-data"),
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())

		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()
		logFile.Close()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		blackboxRunner.Stop()
	})

	It("ignores files in source directory", func() {
		err := ioutil.WriteFile(
			filepath.Join(logDir, "not-a-tag-dir.log"),
			[]byte("some-data"),
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())

		config := buildConfig(logDir)
		blackboxRunner.StartWithConfig(config)

		logFile.WriteString("hello\n")
		logFile.Sync()
		logFile.Close()

		var message *sl.Message
		Eventually(inbox.Messages, "5s").Should(Receive(&message))
		Expect(message.Content).To(ContainSubstring("hello"))
		Expect(message.Content).To(ContainSubstring("test-tag"))
		Expect(message.Content).To(ContainSubstring(Hostname()))

		blackboxRunner.Stop()
	})
})
