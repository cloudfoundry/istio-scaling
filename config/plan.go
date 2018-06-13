package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type TestPlan struct {
	NumApps      int  `json:"number_of_apps"`
	AppInstances int  `json:"app_instances"`
	Cleanup      bool `json:"cleanup"`
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

func (c TestPlan) Validate() error {
	missingProperties := []string{}
	if c.NumApps == 0 {
		missingProperties = append(missingProperties, "number_of_apps")
	}
	if c.AppInstances == 0 {
		missingProperties = append(missingProperties, "app_instances")
	}
	if len(missingProperties) > 0 {
		return errors.New(fmt.Sprintf("Missing required config properties: %s", strings.Join(missingProperties, ", ")))
	}
	return nil
}
