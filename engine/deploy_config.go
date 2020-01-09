package engine

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

type AppRegionConfig struct {
	Name      string
	Enable    bool
	ChartFile string
	Chart     string
	ValueFile string
	Value     interface{}
}

type RegionConfig struct {
	Name              string
	AppRegionConfigs []AppRegionConfig
	StageConfigs      []StageConfig
	ValueFile         string
	Value             interface{}
}


type AppStageConfig struct {
	Name      string
	Enable    bool
	ChartFile string
	Chart     string
	ValueFile string
	Value     interface{}
}

type StageConfig struct {
	Name       string
	AppStageConfigs []AppStageConfig
	EnvConfigs []EnvConfig
	ValueFile  string
	Value      interface{}
}

type AppEnvironmentConfig struct {
	Name      string
	Enable    bool
	ChartFile string
	Chart     string
	ValueFile string
	Value     interface{}
}

type EnvConfig struct {
	Name               string
	AppEnvironmentConfigs []AppEnvironmentConfig
	ValueFile          string
	Value              interface{}
}

func parseRegion(destPath string, pRegion *RegionConfig) error {
	files, err := ioutil.ReadDir(destPath + "/envs")
	if err != nil {
		return err
	}
	
	for _, file := range files {
		filePath := destPath + "/envs/" + file.Name()
		if (file.IsDir()) {
			stage, err := parseStage(file, filePath)
			if err != nil {
				return err
			}
			pRegion.StageConfigs = append(pRegion.StageConfigs, stage)
		} else {
			ext := path.Ext(file.Name())
			if ext == ".yaml" {
				if file.Name() == "values.yaml" {
					pRegion.ValueFile = filePath
					fmt.Println("values.yaml for region ", filePath)
					continue
				}
			}
			pRegion.AppRegionConfigs, err = parseRegionFiles(file, filePath, pRegion.AppRegionConfigs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseStage(stage os.FileInfo, stagePath string) (StageConfig, error) {
	var stageConfig StageConfig
	stageConfig.Name = stage.Name()
	// Look deeper into Stage ....
	stageFiles, err := ioutil.ReadDir(stagePath)
	if err != nil {
		return stageConfig, err
	}
	
	for _, stageFile := range stageFiles {
		filePath := stagePath + "/" + stageFile.Name()
		if (stageFile.IsDir()) {
			environment, err := parseEvnironment(stageFile, filePath)
			if err != nil {
				return stageConfig, err
			}
			stageConfig.EnvConfigs = append(stageConfig.EnvConfigs, environment)
		} else {
			ext := path.Ext(stageFile.Name())
			if ext == ".yaml" {
				if stageFile.Name() == "values.yaml" {
					stageConfig.ValueFile = filePath
					fmt.Println("values.yaml for region ", filePath)
					continue
				}
			}
			stageConfig.AppStageConfigs, err = parseStageFiles(stageFile, filePath, stageConfig.AppStageConfigs)
			if err != nil {
				return stageConfig, err
			}
		}
	}
	
	return stageConfig, nil
}

func parseEvnironment(environment os.FileInfo, environmentPath string) (EnvConfig, error) {
	var environmentConfig EnvConfig
	environmentConfig.Name = environment.Name()
	// Look deeper into Stage ....
	environmentFiles, err := ioutil.ReadDir(environmentPath)
	if err != nil {
		return environmentConfig, err
	}
	
	for _, environmentFile := range environmentFiles {
		filePath := environmentPath + "/" + environmentFile.Name()
		if (environmentFile.IsDir()) {
			// Should not happen
			continue
		} else {
			ext := path.Ext(environmentFile.Name())
			if ext == ".yaml" {
				if environmentFile.Name() == "values.yaml" {
					environmentConfig.ValueFile = filePath
					fmt.Println("values.yaml for region ", filePath)
					continue
				}
			}
			environmentConfig.AppEnvironmentConfigs, err = parseEnvironmentFiles(environmentFile, filePath, environmentConfig.AppEnvironmentConfigs)
			if err != nil {
				return environmentConfig, err
			}
		}
	}
	
	return environmentConfig, nil
}

func parseRegionFiles (file os.FileInfo, filePath string, appRegionConfigs []AppRegionConfig) ([]AppRegionConfig, error) {
	var appC AppRegionConfig
	var found bool
	ext := path.Ext(file.Name())
	if ext == ".yaml" {
		reg := regexp.MustCompile("-values.yaml")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ValueFile = filePath
		fmt.Println("Application Catalog is: ", split[0])
		fmt.Println("Application Catalog Value is: ", file.Name())
	} else if ext == ".txt" {
		reg := regexp.MustCompile("-version.txt")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ChartFile = filePath
		fmt.Println("Application Catalog is: ", split[0])
		fmt.Println("Application Catalog Value is: ", file.Name())
	}
	// Need to parse appCatalogConfigs and see if there is an entry already here.
	for i, a := range appRegionConfigs {
		if a.Name == appC.Name {
			found = true
			if len(appC.ValueFile) > 0 {
				appRegionConfigs[i].ValueFile = appC.ValueFile
			} else if len(appC.ChartFile) > 0 {
				appRegionConfigs[i].ChartFile = appC.ChartFile
			}
			break
		}
	}
	
	if !found {
		appRegionConfigs = append(appRegionConfigs, appC)
	}
	
	return appRegionConfigs, nil
}

func parseStageFiles (file os.FileInfo, filePath string, appStageConfigs []AppStageConfig) ([]AppStageConfig, error) {
	var appC AppStageConfig
	var found bool
	ext := path.Ext(file.Name())
	if ext == ".yaml" {
		reg := regexp.MustCompile("-values.yaml")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ValueFile = filePath
		fmt.Println("Application Catalog is: ", split[0])
		fmt.Println("Application Catalog Value is: ", file.Name())
	} else if ext == ".txt" {
		reg := regexp.MustCompile("-version.txt")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ChartFile = filePath
		fmt.Println("Application Catalog is: ", split[0])
		fmt.Println("Application Catalog Value is: ", file.Name())
	}
	// Need to parse appCatalogConfigs and see if there is an entry already here.
	for i, a := range appStageConfigs {
		if a.Name == appC.Name {
			found = true
			if len(appC.ValueFile) > 0 {
				appStageConfigs[i].ValueFile = appC.ValueFile
			} else if len(appC.ChartFile) > 0 {
				appStageConfigs[i].ChartFile = appC.ChartFile
			}
			break
		}
	}
	
	if !found {
		appStageConfigs = append(appStageConfigs, appC)
	}
	
	return appStageConfigs, nil
}

func parseEnvironmentFiles (file os.FileInfo, filePath string, appEnvironmentConfigs []AppEnvironmentConfig) ([]AppEnvironmentConfig, error) {
	var appC AppEnvironmentConfig
	var found bool
	ext := path.Ext(file.Name())
	if ext == ".yaml" {
		reg := regexp.MustCompile("-values.yaml")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ValueFile = filePath
		fmt.Println("Application Catalog is: ", split[0])
		fmt.Println("Application Catalog Value is: ", file.Name())
	} else if ext == ".txt" {
		reg := regexp.MustCompile("-version.txt")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ChartFile = filePath
		fmt.Println("Application Catalog is: ", split[0])
		fmt.Println("Application Catalog Value is: ", file.Name())
	}
	// Need to parse appCatalogConfigs and see if there is an entry already here.
	for i, a := range appEnvironmentConfigs {
		if a.Name == appC.Name {
			found = true
			if len(appC.ValueFile) > 0 {
				appEnvironmentConfigs[i].ValueFile = appC.ValueFile
			} else if len(appC.ChartFile) > 0 {
				appEnvironmentConfigs[i].ChartFile = appC.ChartFile
			}
			break
		}
	}
	
	if !found {
		appEnvironmentConfigs = append(appEnvironmentConfigs, appC)
	}
	
	return appEnvironmentConfigs, nil
}
