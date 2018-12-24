package scaling_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"code.cloudfoundry.org/istio-scaling/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type Entity struct {
	State   string `json:"state"`
	AppName string `json:"name"`
}

type Metadata struct {
	Guid string `json:"guid"`
}

type Resource struct {
	Entity   Entity   `json:"entity"`
	Metadata Metadata `json:"metadata"`
}

type Response struct {
	Resources []Resource `json:"resources"`
}

func TestIstioScaling(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IstioScaling Suite")
}

var (
	cfg            config.Config
	testPlan       config.TestPlan
	testSetup      *workflowhelpers.ReproducibleTestSuiteSetup
	defaultTimeout = 90 * time.Second
	routeTimeout   = 60 * time.Second
	routesQuota    = -1 // unlimited
)

var _ = BeforeSuite(func() {
	var err error

	configPath := os.Getenv("CONFIG")
	Expect(configPath).NotTo(BeEmpty())
	cfg, err = config.NewConfig(configPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg.Validate()).To(Succeed())

	// change the default quota insteaad of creating a new one
	// Expect(cf.Cf("update-quota", "default", "-r", strconv.Itoa(routesQuota)).Wait(4 * defaultTimeout)).To(Exit(0))
	// Expect(cf.Cf("create-org", cfg.OrgName).Wait(4 * defaultTimeout)).To(Exit(0))
	// Expect(cf.Cf("target", "-o", cfg.OrgName).Wait(4 * defaultTimeout)).To(Exit(0))
	// Expect(cf.Cf("create-space", cfg.SpaceName).Wait(4 * defaultTimeout)).To(Exit(0))
	// Expect(cf.Cf("target", "-o", cfg.OrgName, "-s", cfg.SpaceName).Wait(4 * defaultTimeout)).To(Exit(0))

	// 	testSpace := config.NewSpace(cfg)
	// 	testUser := config.NewUser(cfg)
	// 	adminUser := config.NewAdmin(cfg)
	// 	regularUserCtx := workflowhelpers.NewUserContext(cfg.GetApiEndpoint(), testUser, testSpace, cfg.GetSkipSSLValidation(), defaultTimeout)
	// 	adminUserCtx := workflowhelpers.NewUserContext(cfg.GetApiEndpoint(), adminUser, nil, cfg.GetSkipSSLValidation(), defaultTimeout)
	// 	skipUserCreation := cfg.GetUseExistingUser()
	// 	testSetup = workflowhelpers.NewBaseTestSuiteSetup(cfg, testSpace, testUser, regularUserCtx, adminUserCtx, skipUserCreation)
	// 	testSetup.Setup()

	planPath := os.Getenv("PLAN")
	Expect(planPath).NotTo(BeEmpty())
	testPlan, err = config.NewPlan(planPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(testPlan.Validate()).To(Succeed())

	testSetup = workflowhelpers.NewRunawayAppTestSuiteSetup(cfg)
	testSetup.Setup()

	By(fmt.Sprintf("pushing %d apps", testPlan.NumAppsToPush))
	guaranteePush(testPlan)
})

var _ = AfterSuite(func() {
	if testPlan.Cleanup && testSetup != nil {
		workflowhelpers.AsUser(testSetup.AdminUserContext(), defaultTimeout, func() {
			Expect(cf.Cf("delete-org", "-f", testSetup.GetOrganizationName()).Wait(4 * defaultTimeout)).To(Exit(0))
		})

		testSetup.Teardown()
	}
})

func guaranteePush(testPlan config.TestPlan) {
	pushApps(testPlan.NumAppsToPush, testPlan.NumAppsToCurl, testPlan.Concurrency)
	tryTick := time.Tick(defaultTimeout)
	tries := 1
	runningApps := make(chan int)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case started := <-runningApps:
				if started >= testPlan.NumAppsToCurl {
					return
				}
				unPushedApps := testPlan.NumAppsToCurl - started
				if unPushedApps != 0 {
					fmt.Printf("Started %d apps so far, pushing %d more apps...\n", started, unPushedApps)
					Expect(pushApps(unPushedApps, testPlan.NumAppsToCurl, testPlan.Concurrency)).To(Succeed())
				}
			case <-quit:
				return
			}
		}
	}()

	for range tryTick {
		started := len(startedApps(testPlan.NumAppsToCurl))
		if started >= testPlan.NumAppsToCurl {
			quit <- struct{}{}
			return
		}
		tries += 1
		runningApps <- started
		if tries > 2 {
			quit <- struct{}{}
			return
		}
	}
}

func pushApps(numAppsToPush, totalApps, concurrency int) error {
	started := len(startedApps(testPlan.NumAppsToCurl))
	if started >= totalApps {
		return nil
	}
	sem := make(chan bool, concurrency)
	errs := make(chan error, numAppsToPush)

	for i := 0; i < numAppsToPush; i++ {
		sem <- true
		appName := generator.PrefixedRandomName("SCALING", "APP")
		go func() {
			defer func() { <-sem }()

			err := pushApp(appName)
			if err != nil {
				errs <- err
			}
		}()
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	unstarted := unstartedApps(totalApps)
	if len(unstarted) > 0 {
		err := retryApps(unstarted)
		if err != nil {
			errs <- err
			close(errs)
			return <-errs
		}
	}

	// we don't care about push errors unless we weren't able to
	// successfully restart the unpushed apps
	close(errs)
	return nil
}

func retryApps(unstarted []Resource) error {
	// start unstarted apps
	var delete []string
	for _, u := range unstarted {
		err := startApp(u.Entity.AppName)
		if err != nil {
			delete = append(delete, u.Entity.AppName)
		}
	}

	if len(delete) == 0 {
		return nil
	}

	// delete unstarted apps
	for _, app := range delete {
		err := deleteApp(app)
		if err != nil {
			return err
		}

		appName := generator.PrefixedRandomName("SCALING", "APP")
		// TODO: do something with this err
		pushApp(appName)
	}
	return nil
}

func allApps(appNums int) (resources []Resource) {
	// This is due to https://github.com/cloudfoundry/capi-release/blob/0439fe2157747a7698a5ae09a1f01e034fcaaf9e/jobs/cloud_controller_ng/spec#L708
	maxResultPerPage := 100

	pagination := appNums % maxResultPerPage
	if appNums <= maxResultPerPage {
		resources = append(resources, appsSummary(1, maxResultPerPage)...)
	}
	if pagination != 0 && appNums > maxResultPerPage {
		totalPages := appNums / maxResultPerPage
		for i := 1; i <= totalPages+1; i++ {
			resources = append(resources, appsSummary(i, maxResultPerPage)...)
		}
	}
	if pagination == 0 && appNums > maxResultPerPage {
		totalPages := appNums / maxResultPerPage
		for i := 1; i <= totalPages; i++ {
			resources = append(resources, appsSummary(i, maxResultPerPage)...)
		}
	}
	return resources
}

func startedApps(appNums int) (started []Resource) {
	res := allApps(appNums)
	for _, r := range res {
		if r.Entity.State == "STARTED" {
			started = append(started, r)
		}
	}
	return started
}

func unstartedApps(appNums int) (unstarted []Resource) {
	res := allApps(appNums)
	for _, r := range res {
		if r.Entity.State != "STARTED" {
			unstarted = append(unstarted, r)
		}
	}
	return unstarted
}

func appsSummary(page int, resultPerPage int) []Resource {
	// orgGuid := cf.Cf("org", cfg.OrgName, "--guid").Wait(defaultTimeout).Out.Contents()
	// spaceGuid := cf.Cf("space", cfg.SpaceName, "--guid").Wait(defaultTimeout).Out.Contents()

	// url := fmt.Sprintf("/v2/apps?q=organization_guid:%s&q=space_guid:%s&results-per-page=%d&page=%d", strings.TrimSpace(string(orgGuid)), strings.TrimSpace(string(spaceGuid)), resultPerPage, page)
	// fmt.Printf("calling apps api : cf curl \"%s\"\n", url)
	// bytes, err := exec.Command("cf", "curl", url).CombinedOutput()
	fmt.Printf("calling apps api : cf curl \"/v2/apps?results-per-page=%d&page=%d\"\n", resultPerPage, page)
	bytes, err := exec.Command("cf", "curl", fmt.Sprintf("/v2/apps?results-per-page=%d&page=%d", resultPerPage, page)).CombinedOutput()
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}
	var resp Response
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}

	return resp.Resources
}

func deleteApp(name string) error {
	fmt.Printf("deleting unstarted app: cf delete %s\n", name)
	_, err := exec.Command("cf", "delete", "-f", "-r", name).CombinedOutput()
	return err
}

func startApp(name string) error {
	fmt.Printf("starting Stopped app: cf start %s\n", name)
	_, err := exec.Command("cf", "start", name).CombinedOutput()
	return err
}

func pushApp(appName string) error {
	fmt.Printf("pushing: %s \n", appName)
	bytes, err := exec.Command("cf",
		"push", appName,
		"-s", "cflinuxfs3",
		"-d", cfg.IstioDomain,
		"--droplet", appDropletPath,
		"-i", fmt.Sprintf("%d", testPlan.AppInstances),
		"-m", testPlan.AppMemSize,
		"-k", "75M").CombinedOutput()
	fmt.Printf("output: %s\n", string(bytes))
	return err
}
