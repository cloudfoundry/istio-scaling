package config

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type Space struct {
	cfg Config
}

func NewSpace(cfg Config) *Space {
	return &Space{cfg: cfg}
}

func (s *Space) Create() {
	Expect(cf.Cf(
		"create-quota",
		s.QuotaName(),
		"-m", workflowhelpers.RUNAWAY_QUOTA_MEM_LIMIT,
		"-i", "-1",
		"-r", "-1",
		"-a", "-1",
		"-s", "-1",
		"--reserved-route-ports", "20",
		"--allow-paid-service-plans").Wait(5 * defaultTimeout)).To(Exit(0))
	Expect(cf.Cf("create-org", s.OrganizationName()).Wait(5 * defaultTimeout)).To(Exit(0))
	Expect(cf.Cf("set-quota", s.OrganizationName(), s.QuotaName()).Wait(5 * defaultTimeout)).To(Exit(0))
	Expect(cf.Cf("create-space", "-o", s.OrganizationName(), s.SpaceName()).Wait(5 * defaultTimeout)).To(Exit(0))
}

func (s *Space) Destroy() {}

func (s *Space) ShouldRemain() bool { return true }

func (s *Space) OrganizationName() string {
	return s.cfg.GetExistingOrganization()
}

func (s *Space) SpaceName() string {
	return s.cfg.GetExistingSpace()
}

func (s *Space) QuotaName() string {
	return "scaling-test-quota"
}
