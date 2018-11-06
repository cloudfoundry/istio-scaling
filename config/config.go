package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	CFSystemDomain string `json:"cf_system_domain"`
	IstioDomain    string `json:"cf_istio_domain"`
	AdminUser      string `json:"cf_admin_user"`
	AdminPassword  string `json:"cf_admin_password"`
	OrgName        string `json:"cf_org_name"`
	SpaceName      string `json:"cf_space_name"`
}

func NewConfig(path string) (Config, error) {
	var config Config

	configFile, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	return config, err
}

func (c Config) Validate() error {
	missingProperties := []string{}
	if c.IstioDomain == "" {
		missingProperties = append(missingProperties, "cf_istio_domain")
	}
	if c.CFSystemDomain == "" {
		missingProperties = append(missingProperties, "cf_system_domain")
	}
	if c.AdminUser == "" {
		missingProperties = append(missingProperties, "cf_admin_user")
	}
	if c.AdminPassword == "" {
		missingProperties = append(missingProperties, "cf_admin_password")
	}
	if len(missingProperties) > 0 {
		return errors.New(fmt.Sprintf("Missing required config properties: %s", strings.Join(missingProperties, ", ")))
	}
	return nil
}

func (c Config) GetApiEndpoint() string {
	return "api." + c.CFSystemDomain
}

func (c Config) GetAdminPassword() string {
	return c.AdminPassword
}

func (c Config) GetAdminUser() string {
	return c.AdminUser
}

func (c Config) GetConfigurableTestPassword() string            { return "" }
func (c Config) GetPersistentAppOrg() string                    { return "" }
func (c Config) GetPersistentAppQuotaName() string              { return "" }
func (c Config) GetPersistentAppSpace() string                  { return "" }
func (c Config) GetScaledTimeout(d time.Duration) time.Duration { return d }
func (c Config) GetExistingUser() string                        { return "" }
func (c Config) GetExistingUserPassword() string                { return "" }
func (c Config) GetShouldKeepUser() bool                        { return false }
func (c Config) GetUseExistingUser() bool                       { return false }
func (c Config) GetExistingOrganization() string                { return c.OrgName }
func (c Config) GetExistingSpace() string                       { return c.SpaceName }
func (c Config) GetSkipSSLValidation() bool                     { return true }
func (c Config) GetNamePrefix() string                          { return "SCALING" }

func (c Config) GetUseExistingOrganization() bool {
	if c.OrgName != "" {
		return true
	}
	return false
}

func (c Config) GetUseExistingSpace() bool {
	if c.SpaceName != "" {
		return true
	}
	return false
}
