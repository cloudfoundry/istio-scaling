package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type TestPlan struct {
	NumAppsToPush int    `json:"number_of_apps_to_push"`
	NumAppsToCurl int    `json:"number_of_apps_to_curl"`
	AppMemSize    string `json:"app_mem_size" default:"16M"`
	AppInstances  int    `json:"app_instances"`
	Concurrency   int    `json:"app_push_concurrency"`
	Cleanup       bool   `json:"cleanup"`
}

func NewPlan(path string) (TestPlan, error) {
	var config TestPlan

	configFile, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	return config, err
}

func (c *TestPlan) Validate() error {
	missingProperties := []string{}
	if c.NumAppsToPush == 0 {
		missingProperties = append(missingProperties, "number_of_apps_to_push")
	}
	if c.NumAppsToCurl == 0 {
		c.NumAppsToCurl = c.NumAppsToPush
	}
	if c.NumAppsToCurl < c.NumAppsToPush {
		return errors.New(("number_of_apps_to_curl must be >= number_of_apps_to_push"))
	}
	if c.AppInstances == 0 {
		missingProperties = append(missingProperties, "app_instances")
	}
	if c.Concurrency == 0 {
		c.Concurrency = 16
	}
	if len(missingProperties) > 0 {
		return fmt.Errorf("Missing required config properties: %s", strings.Join(missingProperties, ", "))
	}
	return nil
}
