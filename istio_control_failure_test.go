package scaling_test

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Control Plane Failure", func() {
	var deployment boshdir.Deployment
	var instance boshdir.AllOrInstanceGroupOrInstanceSlug
	var hostname string

	Context("when copilot fails with data", func() {
		BeforeEach(func() {
			var options = struct {
				DirectorCA       string
				DirectorUser     string
				DirectorPassword string
				DirectorUAAURL   string
				DirectorURL      string
				DeploymentName   string
			}{
				DirectorCA:       os.Getenv("BOSH_CA_CERT"),
				DirectorUser:     os.Getenv("BOSH_CLIENT"),
				DirectorPassword: os.Getenv("BOSH_CLIENT_SECRET"),
				DirectorUAAURL:   fmt.Sprintf("%s:8443", strings.TrimSuffix(os.Getenv("BOSH_ENVIRONMENT"), ":25555")),
				DirectorURL:      os.Getenv("BOSH_ENVIRONMENT"),
				DeploymentName:   os.Getenv("BOSH_DEPLOYMENT"),
			}

			logger := boshlog.NewLogger(boshlog.LevelError)
			uaaFactory := boshuaa.NewFactory(logger)

			uaaCfg, err := boshuaa.NewConfigFromURL(options.DirectorUAAURL)
			Expect(err).NotTo(HaveOccurred())

			uaaCfg.Client = options.DirectorUser
			uaaCfg.ClientSecret = options.DirectorPassword
			uaaCfg.CACert = options.DirectorCA

			uaa, err := uaaFactory.New(uaaCfg)
			Expect(err).NotTo(HaveOccurred())

			directorFactory := boshdir.NewFactory(logger)

			directorCfg, err := boshdir.NewConfigFromURL(options.DirectorURL)
			Expect(err).NotTo(HaveOccurred())

			directorCfg.CACert = options.DirectorCA
			directorCfg.TokenFunc = boshuaa.NewClientTokenSession(uaa).TokenFunc
			director, err := directorFactory.New(directorCfg, boshdir.NewNoopTaskReporter(), boshdir.NewNoopFileReporter())
			Expect(err).NotTo(HaveOccurred())

			deployment, err = director.FindDeployment(options.DeploymentName)
			Expect(err).NotTo(HaveOccurred())

			hostname = "failure-test-route"
			instance = boshdir.NewAllOrInstanceGroupOrInstanceSlug("istio-control", "0")
		})

		AfterEach(func() {
			cf.Cf("delete-route", cfg.IstioDomain, "--hostname", hostname)
		})

		It("returns to normal operation after some time", func() {
			By("stopping the istio control vm")
			err := deployment.Stop(instance, boshdir.StopOpts{
				SkipDrain: true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(err).NotTo(HaveOccurred())
			Eventually(func() (string, error) {
				vms, err := deployment.VMInfos()
				Expect(err).NotTo(HaveOccurred())

				for _, vm := range vms {
					if vm.JobName == "istio-control" {
						return vm.ProcessState, nil
					}
				}

				return "", errors.New("could not find istio vm")
			}, 10*time.Minute, "1m").Should(Equal("stopped"))

			By("mapping a new route")
			apps := allApps(testPlan.NumApps)
			app := apps[0].Entity.AppName

			cf.Cf("map-route", app, cfg.IstioDomain, "--hostname", hostname)

			appURL := fmt.Sprintf("http://%s.%s", hostname, cfg.IstioDomain)

			By("confirming that the new route has not been synced")
			resp, err := http.Get(appURL)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))

			By("restarting the istio control vm")
			deployment.Start(instance, boshdir.StartOpts{})
			Eventually(func() (string, error) {
				vms, err := deployment.VMInfos()
				Expect(err).NotTo(HaveOccurred())

				for _, vm := range vms {
					if vm.JobName == "istio-control" {
						return vm.ProcessState, nil
					}
				}

				return "", errors.New("could not find istio vm")
			}, 10*time.Minute, "30s").Should(Equal("running"))

			By("confirming that the new route is synced")
			Eventually(func() (int, error) {
				var err error
				resp, err = http.Get(appURL)
				if err != nil {
					return http.StatusTeapot, err
				}

				return resp.StatusCode, nil
			}, defaultTimeout, "30s").Should(Equal(http.StatusOK))
		})
	})
})
