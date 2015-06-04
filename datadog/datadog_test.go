package datadog_test

import (
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/concourse/blackbox/datadog"
)

type request struct {
	Series datadog.Series `json:"series"`
}

var _ = Describe("Datadog", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		datadog.APIURL = server.URL()
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Context("when everything's great", func() {
		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/series", "api_key=api-key"),
				func(w http.ResponseWriter, r *http.Request) {
					var request request
					Ω(json.NewDecoder(r.Body).Decode(&request)).Should(Succeed())
					metric := request.Series[0]

					Ω(metric.Name).Should(Equal("memory.limit"))
					Ω(metric.Host).Should(Equal("web-0"))
					Ω(metric.Tags).Should(ConsistOf("application:atc"))

					Ω(metric.Points[0].Timestamp).Should(BeTemporally("~", time.Now(), time.Second))
					Ω(metric.Points[0].Value).Should(BeNumerically("~", 4.52, 0.01))

					Ω(metric.Points[1].Timestamp).Should(BeTemporally("~", time.Now(), time.Second))
					Ω(metric.Points[1].Value).Should(BeNumerically("~", 23.22, 0.01))

					Ω(metric.Points[2].Timestamp).Should(BeTemporally("~", time.Now(), time.Second))
					Ω(metric.Points[2].Value).Should(BeNumerically("~", 23.25, 0.01))
				},
				ghttp.RespondWith(http.StatusAccepted, "{}"),
			))
		})

		It("works", func() {
			client := datadog.NewClient("api-key")

			err := client.PublishSeries(datadog.Series{
				{
					Name: "memory.limit",
					Points: []datadog.Point{
						{time.Now(), 4.52},
						{time.Now(), 23.22},
						{time.Now(), 23.25},
					},
					Host: "web-0",
					Tags: []string{"application:atc"},
				},
			})

			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("when the server does not respond", func() {
		BeforeEach(func() {
			server.Close()
			server = nil
		})

		It("returns an error", func() {
			client := datadog.NewClient("api-key")

			err := client.PublishSeries(datadog.Series{})
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("when the server does not respond with 202", func() {
		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/series", "api_key=api-key"),
				ghttp.RespondWith(http.StatusInternalServerError, "{}"),
			))
		})

		It("returns an error", func() {
			client := datadog.NewClient("api-key")

			err := client.PublishSeries(datadog.Series{})
			Ω(err).Should(HaveOccurred())
		})
	})
})
