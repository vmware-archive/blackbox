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
			Expect(err).NotTo(HaveOccurred())
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
		Expect(err).NotTo(HaveOccurred())

		seen := map[string]float32{}

		expvars.Walk(func(path string, value float32) {
			seen[path] = value
		})

		// some random keys
		Expect(seen).To(HaveKeyWithValue("memstats.Alloc", float32(202208)))
		Expect(seen).To(HaveKeyWithValue("memstats.HeapReleased", float32(0)))
	})

	It("lets people get the size of the expvars", func() {
		expvars, err := fetcher.Fetch()
		Expect(err).NotTo(HaveOccurred())

		Expect(expvars.Size()).To(Equal(25))
	})

	Context("when the server is down", func() {
		BeforeEach(func() {
			server.Close()
			server = nil
		})

		It("returns an error", func() {
			_, err := fetcher.Fetch()
			Expect(err).To(HaveOccurred())
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
			Expect(err).To(HaveOccurred())
		})
	})
})
