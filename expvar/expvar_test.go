package expvar_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"

	"github.com/concourse/blackbox/expvar"
)

var _ = Describe("Expvar", func() {
	var (
		server  *ghttp.Server
		fetcher expvar.Fetcher
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		server.RouteToHandler("GET", "/debug/vars", func(w http.ResponseWriter, r *http.Request) {
			contents, err := ioutil.ReadFile(filepath.Join("fixtures", "expvar.json"))
			Ω(err).ShouldNot(HaveOccurred())
			w.Write(contents)
		})

		fetcher = expvar.NewFetcher(server.URL() + "/debug/vars")
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	It("fetches expvar from a server", func() {
		expvars, err := fetcher.Fetch()
		Ω(err).ShouldNot(HaveOccurred())

		seen := map[string]float32{}

		expvars.Walk(func(path string, value float32) {
			seen[path] = value
		})

		// some random keys
		Ω(seen).Should(HaveKeyWithValue("memstats.Alloc", float32(202208)))
		Ω(seen).Should(HaveKeyWithValue("memstats.HeapReleased", float32(0)))
	})

	Context("when the server is down", func() {
		BeforeEach(func() {
			server.Close()
			server = nil
		})

		It("returns an error", func() {
			_, err := fetcher.Fetch()
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("when the server serves some non-json", func() {
		BeforeEach(func() {
			server.RouteToHandler("GET", "/debug/vars", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "{{{{}}")
			})
		})

		It("returns an error", func() {
			_, err := fetcher.Fetch()
			Ω(err).Should(HaveOccurred())
		})
	})
})
