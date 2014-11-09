package integration_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
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

func Hostname() string {
	hostname, err := os.Hostname()
	Ω(err).ShouldNot(HaveOccurred())
	return hostname
}
