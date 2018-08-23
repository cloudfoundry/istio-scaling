package scaling_test

import (
	"os"
	"testing"
	"time"

	"code.cloudfoundry.org/istio-scaling/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func TestIstioScaling(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IstioScaling Suite")
}

var (
	cfg            config.Config
	testPlan       config.TestPlan
	testSetup      *workflowhelpers.ReproducibleTestSuiteSetup
	defaultTimeout = 60 * time.Second
)

var _ = BeforeSuite(func() {
	var err error

	configPath := os.Getenv("CONFIG")
	Expect(configPath).NotTo(BeEmpty())
	cfg, err = config.NewConfig(configPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg.Validate()).To(Succeed())

	planPath := os.Getenv("PLAN")
	Expect(planPath).NotTo(BeEmpty())
	testPlan, err = config.NewPlan(planPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(testPlan.Validate()).To(Succeed())

	testSetup = workflowhelpers.NewRunawayAppTestSuiteSetup(cfg)
	testSetup.Setup()
})

var _ = AfterEach(func() {
	if testPlan.Cleanup && testSetup != nil {
		workflowhelpers.AsUser(testSetup.AdminUserContext(), defaultTimeout, func() {
			Expect(cf.Cf("delete-org", "-f", testSetup.GetOrganizationName()).Wait(defaultTimeout)).To(Exit(0))
		})

		testSetup.Teardown()
	}
})
