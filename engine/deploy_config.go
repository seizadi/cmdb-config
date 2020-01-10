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

type TierConfig struct {
	Name        string
	AppConfigs  []AppConfig
	ValueFile   string
	Value       interface{}
	TierConfigs []TierConfig
}

func parseTier(tier os.FileInfo, tierPath string, levels []int) (TierConfig, error) {
	tierLevel := levels[1:]
	var tierConfig TierConfig
	tierConfig.Name = tier.Name()
	// Look deeper into Stage ....
	files, err := ioutil.ReadDir(tierPath)
	if err != nil {
		return tierConfig, err
	}
	
	for _, file := range files {
		filePath := tierPath + "/" + file.Name()
		if len(tierLevel) > 0 && file.IsDir() {
			tier, err := parseTier(file, filePath, tierLevel)
			if err != nil {
				return tierConfig, err
			}
			tierConfig.TierConfigs = append(tierConfig.TierConfigs, tier)
		} else {
			ext := path.Ext(file.Name())
			if ext == ".yaml" {
				if file.Name() == "values.yaml" {
					tierConfig.ValueFile = filePath
					continue
				}
			}
			tierConfig.AppConfigs, err = parseAppFiles(file, filePath, tierConfig.AppConfigs)
			if err != nil {
				return tierConfig, err
			}
		}
	}
	
	return tierConfig, nil
}

func parseAppFiles (file os.FileInfo, filePath string, appConfigs []AppConfig) ([]AppConfig, error) {
	var found bool
	appC := AppConfig { Enable: true}
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
		if err !=nil {
			return appConfigs, err
		}
		chartStr := strings.TrimSpace(string(content))
		appC.Chart = chartStr
		if (chartStr == "dnd") {
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
