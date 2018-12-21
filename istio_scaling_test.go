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

func buildDatadogResponse(number int, metric string, int timestamp) string {
	return fmt.Sprintf(`
	{
		"series" :
			[{
				"metric":"scaling_test.wip.%s",
			  "points":[[%d, %d]],
			  "type":"count",
			  "tags":["deployment:%s"]
			}]
	}`,
		metric,
		timestamp,
		number,
		cfg.CFSystemDomain,
	)
}

func sendResultToDatadog(numberSuccessfulCurls int, totalCurls int) {
	environment := getValidDatadogName(strings.Split(cfg.CFSystemDomain, ".")[0])
	timestamp := time.Now()

	successData := buildDatadogResponse(numberSuccessfulCurls, "success", timestamp)
	totalData := buildDatadogResponse(totalCurls, "total", timestamp)

	url := fmt.Sprintf("https://app.datadoghq.com/api/v1/series?api_key=%s", cfg.DatadogApiKey)
	client := http.DefaultClient()

	resp, err := client.Post(url, "application/json", strings.NewReader(successData))
	Expect(err).NotTo(HaveOccurred())

	resp, err := client.Post(url, "application/json", strings.NewReader(totalData))
	Expect(err).NotTo(HaveOccurred())

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
				sendResultToDatadog(appsUpCount, testPlan.NumAppsToCurl)
			}
		})
	})
})

func getValidDatadogName(name string) string {
	return strings.Replace(name, "-", "_", -1)
}
