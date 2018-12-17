package scaling_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var appDropletPath = "assets/hello-golang.tgz"

var _ = Describe("Istio scaling", func() {
	Context("when pushing multiple apps", func() {
		It("checks responses", func() {
			apps := allApps(testPlan.NumAppsToCurl)
			Expect(len(apps)).Should(Equal(testPlan.NumAppsToCurl))

			appsUpCount := 0
			for i, app := range apps {
				appURL := fmt.Sprintf("http://%s.%s", app.Entity.AppName, cfg.IstioDomain)
				By(fmt.Sprintf("%d -- send request to app %s", i, appURL))
				var resp *http.Response
				resp, err := http.Get(appURL)
				if err != nil {
					fmt.Printf("STEP: %d -- failed with error: %+v\n", i, err)
				} else if resp.StatusCode != http.StatusOK {
					fmt.Printf("STEP: %d -- failed with status: %d\n", i, resp.StatusCode)
				} else {
					appsUpCount += 1
				}
			}

			fmt.Println("RESULTS:")
			fmt.Printf("  %d out of %d curls successful\n", appsUpCount, testPlan.NumAppsToCurl)

			environment := getValidDatadogName(strings.Split(cfg.CFSystemDomain, ".")[0])
			metric := fmt.Sprintf("%s.scale.AppsUp", environment)

			data := fmt.Sprintf(`{ "series" :
			           [{"metric":"%s",
			            "points":[[$(date +%%s), %d]],
			            "type":"gauge",
			            "tags":["deployment:%s"]
			          }]
			        }`, metric, appsUpCount, cfg.CFSystemDomain)
			b, err := exec.Command("curl",
				"-f",
				"-X", "POST",
				"-H", "Content-type: application/json",
				"-d", data,
				fmt.Sprintf("https://app.datadoghq.com/api/v1/series?api_key=%s", cfg.DatadogApiKey)).CombinedOutput()
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(bytes.Contains(b, []byte(`{"status": "ok"}`))).Should(BeTrue())
			fmt.Printf("Metric %s sent successfully\n", metric)
		})
	})
})

func getValidDatadogName(name string) string {
	return strings.Replace(name, "-", "_", -1)
}
