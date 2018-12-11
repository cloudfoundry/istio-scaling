package scaling_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var appDropletPath = "assets/hello-golang.tgz"

var _ = Describe("Istio scaling", func() {
	Context("when pushing multiple apps", func() {
		It("checks responses", func() {
			apps := allApps(testPlan.NumAppsToCurl)
			errCount := 0
			for i, app := range apps {
				appURL := fmt.Sprintf("http://%s.%s", app.Entity.AppName, cfg.IstioDomain)
				By(fmt.Sprintf("%d -- send request to app %s", i, appURL))
				var resp *http.Response
				resp, err := http.Get(appURL)
				if err != nil {
					errCount += 1
					fmt.Printf("STEP: %d -- failed with error: %+v\n", i, err)
				} else if resp.StatusCode != http.StatusOK {
					errCount += 1
					fmt.Printf("STEP: %d -- failed with status: %d\n", i, resp.StatusCode)
				}
			}

			errThreshold := (100 - testPlan.PassingThreshold) / 100
			errRatio := float32(errCount) / float32(testPlan.NumAppsToCurl)

			fmt.Println("RESULTS:")
			fmt.Printf("  %d error(s) out of %d curls\n", errCount, testPlan.NumAppsToCurl)
			fmt.Printf("  Error Threshold: %f (%f%% Passing Threshold)\n", errThreshold, testPlan.PassingThreshold)
			fmt.Printf("  Error Actual: %f\n", errRatio)

			Expect(errRatio).Should(BeNumerically("<=", errThreshold))
		})
	})
})
