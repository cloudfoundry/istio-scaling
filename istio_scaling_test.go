package scaling_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	appNames    []string
	appNameLock sync.Mutex
)

var _ = Describe("Istio scaling", func() {

	Context("when pushing multiple apps", func() {

		BeforeEach(func() {
			By(fmt.Sprintf("pushing %d apps", testPlan.NumApps))
			completed := make(chan struct{}, testPlan.NumApps)
			for i := 0; i < testPlan.NumApps; i++ {
				go pushApp(completed)
			}
			Eventually(func() int {
				return len(completed)
			}, defaultTimeout).Should(Equal(testPlan.NumApps))
		})

		It("checks responses", func() {
			for _, appName := range appNames {
				appURL := fmt.Sprintf("http://%s.%s", appName, cfg.IstioDomain)
				By(fmt.Sprintf("send request to app %s", appURL))
				resp, err := http.Get(appURL)
				Expect(err).ToNot(HaveOccurred())

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())

				type AppResponse struct {
					Greeting string `json:"greeting"`
				}

				var appResp AppResponse
				err = json.Unmarshal(body, &appResp)
				Expect(err).ToNot(HaveOccurred())
				Expect(appResp.Greeting).To(Equal("hello"))
			}
		})
	})
})

func pushApp(completed chan struct{}) {
	appName := generator.PrefixedRandomName("SCALING", "APP")
	appNameLock.Lock()
	appNames = append(appNames, appName)
	appNameLock.Unlock()
	Expect(cf.Cf("push", appName,
		"-d", cfg.IstioDomain,
		"-i", fmt.Sprintf("\"%d\"", testPlan.AppInstances),
		"-b", "binary_buildpack",
		"-c", "./"+appBinary,
	).Wait(defaultTimeout)).To(Exit(0))

	appURL := fmt.Sprintf("http://%s.%s", appName, cfg.IstioDomain)
	Eventually(func() (bool, error) {
		statusCode, err := getStatusCode(appURL)
		if err == nil {
			if statusCode == http.StatusOK {
				completed <- struct{}{}
				return true, nil
			}
		}
		return false, err
	}, defaultTimeout).Should(BeTrue())
}

func getStatusCode(appURL string) (int, error) {
	resp, err := http.Get(appURL)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}
