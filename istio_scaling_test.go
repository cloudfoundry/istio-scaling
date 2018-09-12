package scaling_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var appDropletPath = "assets/hello-golang.tgz"

var _ = Describe("Istio scaling", func() {
	Context("when pushing multiple apps", func() {
		It("checks responses", func() {
			apps := allApps(testPlan.NumApps)
			for i, app := range apps {
				appURL := fmt.Sprintf("http://%s.%s", app.Entity.AppName, cfg.IstioDomain)
				By(fmt.Sprintf("%d -- send request to app %s", i, appURL))
				var resp *http.Response
				Eventually(func() (int, error) {
					var err error
					resp, err = http.Get(appURL)
					if err != nil {
						return http.StatusTeapot, err
					}

					return resp.StatusCode, nil
				}, routeTimeout, "2s").Should(Equal(http.StatusOK))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(body)).To(ContainSubstring("hello"))
			}
		})
	})
})
