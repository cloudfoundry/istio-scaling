package scaling_test

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var appDropletPath = "assets/hello-golang.tgz"

func buildDatadogResponse(number int, metric string, timestamp time.Time) string {
	environmentName := getValidDatadogName(strings.Split(cfg.CFSystemDomain, ".")[0])
	return fmt.Sprintf(`
	{
		"series" :
			[{
			  "metric":"istio_scaling_test.%s",
			  "points":[[%d, %d]],
			  "type":"count",
			  "tags":["deployment:%s"]
			}]
	}`,
		metric,
		timestamp.Unix(),
		number,
		environmentName,
	)
}

func sendResultToDatadog(numberSuccessfulCurls int, totalCurls int) {
	timestamp := time.Now()

	successData := buildDatadogResponse(numberSuccessfulCurls, "success", timestamp)
	totalData := buildDatadogResponse(totalCurls, "total", timestamp)

	url := fmt.Sprintf("https://app.datadoghq.com/api/v1/series?api_key=%s", cfg.DatadogApiKey)
	client := http.DefaultClient

	_, err := client.Post(url, "application/json", strings.NewReader(successData))
	Expect(err).NotTo(HaveOccurred())

	_, err = client.Post(url, "application/json", strings.NewReader(totalData))
	Expect(err).NotTo(HaveOccurred())
	fmt.Println("Results sent to datadog!")
}

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

			if cfg.DatadogApiKey != "" {
				fmt.Println("Sending results to datadog")
				sendResultToDatadog(appsUpCount, testPlan.NumAppsToCurl)
			}
		})
	})
})

func getValidDatadogName(name string) string {
	return strings.Replace(name, "-", "_", -1)
}
