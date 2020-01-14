package engine

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

type AppConfig struct {
	Name      string
	Enable    bool
	ChartFile string
	Chart     string
	ValueFile string
	Value     interface{}
}

type LifecycleConfig struct {
	Name             string
	AppConfigs       []AppConfig
	ValueFile        string
	Value            interface{}
	LifecycleConfigs []LifecycleConfig
}

func parseLifecycle(lifecycle os.FileInfo, lifecyclePath string) (LifecycleConfig, error) {
	var lifecycleConfig LifecycleConfig
	lifecycleConfig.Name = lifecycle.Name()
	// Look deeper into Stage ....
	files, err := ioutil.ReadDir(lifecyclePath)
	if err != nil {
		return lifecycleConfig, err
	}

	for _, file := range files {
		filePath := lifecyclePath + "/" + file.Name()
		if file.IsDir() {
			lifecycle, err := parseLifecycle(file, filePath)
			if err != nil {
				return lifecycleConfig, err
			}
			lifecycleConfig.LifecycleConfigs = append(lifecycleConfig.LifecycleConfigs, lifecycle)
		} else {
			ext := path.Ext(file.Name())
			if ext == ".yaml" {
				if file.Name() == "values.yaml" {
					lifecycleConfig.ValueFile = filePath
					continue
				}
			}
			lifecycleConfig.AppConfigs, err = parseAppFiles(file, filePath, lifecycleConfig.AppConfigs)
			if err != nil {
				return lifecycleConfig, err
			}
		}
	}

	return lifecycleConfig, nil
}

func parseAppFiles(file os.FileInfo, filePath string, appConfigs []AppConfig) ([]AppConfig, error) {
	var found bool
	appC := AppConfig{Enable: true}
	ext := path.Ext(file.Name())
	if ext == ".yaml" {
		reg := regexp.MustCompile("-values.yaml")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ValueFile = filePath
	} else if ext == ".txt" {
		reg := regexp.MustCompile("-version.txt")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ChartFile = filePath
		content, err := ioutil.ReadFile(appC.ChartFile)
		if err != nil {
			return appConfigs, err
		}
		chartStr := strings.TrimSpace(string(content))
		appC.Chart = chartStr
		if chartStr == "dnd" {
			appC.Enable = false
		} else {
			appC.Enable = true
		}
	}
	// Need to parse appCatalogConfigs and see if there is an entry already here.
	for i, a := range appConfigs {
		if a.Name == appC.Name {
			found = true
			if len(appC.ValueFile) > 0 {
				appConfigs[i].ValueFile = appC.ValueFile
			} else if len(appC.ChartFile) > 0 {
				appConfigs[i].Enable = appC.Enable
				appConfigs[i].ChartFile = appC.ChartFile
				appConfigs[i].Chart = appC.Chart
			}
			break
		}
	}

	if !found {
		appConfigs = append(appConfigs, appC)
	}

	return appConfigs, nil
}
