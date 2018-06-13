package scaling_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	appNames       []string
	appNameLock    sync.Mutex
	appDropletPath = "assets/dora-droplet.tar.gz"
)

var _ = Describe("Istio scaling", func() {

	Context("when pushing multiple apps", func() {

		BeforeEach(func() {
			By(fmt.Sprintf("pushing %d apps", testPlan.NumApps))
			var wg sync.WaitGroup
			wg.Add(testPlan.NumApps)

			for i := 0; i < testPlan.NumApps; i++ {
				go func() {
					defer wg.Done()
					for {
						status, appName := pushApp()
						if status != 0 {
							cf.Cf("delete", appName, "-f", "-r").Wait(defaultTimeout)
							time.Sleep(2 * time.Second)

							continue
						}

						return
					}
				}()
			}

			fmt.Println("waiting for all apps to be up")
			wg.Wait()
		})

		It("checks responses", func() {
			for _, appName := range appNames {
				appURL := fmt.Sprintf("http://%s.%s", appName, cfg.IstioDomain)
				By(fmt.Sprintf("send request to app %s", appURL))

				var resp *http.Response
				Eventually(func() (int, error) {
					var err error
					resp, err = http.Get(appURL)
					if err != nil {
						return http.StatusTeapot, err
					}

					return resp.StatusCode, nil
				}, defaultTimeout).Should(Equal(http.StatusOK))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(body)).To(Equal("Hi, I'm Dora!"))
			}
		})
	})
})

func pushApp() (int, string) {
	appName := generator.PrefixedRandomName("SCALING", "APP")
	appNameLock.Lock()
	appNames = append(appNames, appName)
	appNameLock.Unlock()
	statusCode := cf.Cf("push", appName,
		"-d", cfg.IstioDomain,
		"--droplet", appDropletPath,
		"-i", fmt.Sprintf("\"%d\"", testPlan.AppInstances),
		"-m", "16M",
		"-k", "75M",
	).Wait(defaultTimeout).ExitCode()

	return statusCode, appName
}
