package integration

import (
	"errors"
	"io/ioutil"
	"log"
	"time"

	dfakes "github.com/concourse/blackbox/datadog/fakes"
	efakes "github.com/concourse/blackbox/expvar/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"github.com/concourse/blackbox"
	"github.com/concourse/blackbox/expvar"
)

var _ = Describe("Emitter", func() {
	var (
		process ifrit.Process

		fakeFetcher *efakes.FakeFetcher
		fakeDatadog *dfakes.FakeClient
	)

	BeforeEach(func() {
		fakeFetcher = &efakes.FakeFetcher{}
		fakeDatadog = &dfakes.FakeClient{}
	})

	JustBeforeEach(func() {
		process = ginkgomon.Invoke(blackbox.NewEmitter(
			fakeDatadog,
			fakeFetcher,
			time.Second,
			"an-amazing-host.local",
			[]string{"some", "great", "tags"},
		))
	})

	AfterEach(func() {
		ginkgomon.Interrupt(process)
	})

	Context("when everything is great", func() {
		BeforeEach(func() {
			expvars := expvar.Expvars{
				"memory": map[string]float32{
					"limit": 3.23,
				},
			}
			fakeFetcher.FetchReturns(expvars, nil)
		})

		It("sends the things to datadog", func() {
			Eventually(fakeDatadog.PublishSeriesCallCount).Should(Equal(1))

			emittedSeries := fakeDatadog.PublishSeriesArgsForCall(0)
			metric := emittedSeries[0]

			Ω(metric.Name).Should(Equal("memory.limit"))
			Ω(metric.Host).Should(Equal("an-amazing-host.local"))
			Ω(metric.Tags).Should(ConsistOf("some", "great", "tags"))

			point := metric.Points[0]
			Ω(point.Timestamp).Should(BeTemporally("~", time.Now()))
			Ω(point.Value).Should(BeNumerically("~", 3.23, 0.0001))
		})
	})

	Context("when fetching expvars fails", func() {
		BeforeEach(func() {
			log.SetOutput(ioutil.Discard)

			error := errors.New("disaster")
			fakeFetcher.FetchReturns(expvar.Expvars{}, error)
		})

		It("does not send anything to datadog", func() {
			Consistently(fakeDatadog.PublishSeriesCallCount).Should(Equal(0))
		})
	})
})
