package engine

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
	BuildPath        string // TODO - Replace with server side logic to compute build tree
}

func parseLifecycle(lifecycle os.FileInfo, lifecyclePath string) (LifecycleConfig, error) {
	var lifecycleConfig LifecycleConfig
	lifecycleConfig.Name = lifecycle.Name()
	lifecycleConfig.BuildPath = getBuildPath(lifecyclePath)
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
			lifecycleConfig.AppConfigs, err = parseAppFiles(file, filePath, lifecycleConfig.AppConfigs, false)
			if err != nil {
				return lifecycleConfig, err
			}
		}
	}

	return lifecycleConfig, nil
}

func getBuildPath(lifecyclePath string) string {
	files := strings.SplitAfter(lifecyclePath, "/envs")
	dir, _ := path.Split(files[0])
	buildPath := path.Join(dir, "build", files[1])
	return buildPath
}

func parseAppFiles(file os.FileInfo, filePath string, appConfigs []AppConfig, envFlag bool) ([]AppConfig, error) {
	var found bool
	appC := AppConfig{Enable: true}
	ext := path.Ext(file.Name())
	if ext == ".yaml" {
		var reg *regexp.Regexp
		if envFlag == false {
			reg = regexp.MustCompile("-values.yaml")
		} else {
			reg = regexp.MustCompile(".yaml")
		}
		split := reg.Split(file.Name(), -1)
		if len(split) > 1 { // Ignore if it doesn't have valid extension
			appC.Name = split[0]
			appC.ValueFile = filePath
		}
	} else if ext == ".txt" {
		reg := regexp.MustCompile("-version.txt")
		split := reg.Split(file.Name(), -1)
		appC.Name = split[0]
		appC.ChartFile = filePath
		// We have symlink used for the files now!
		destFilePath := filePath
		if file.Mode()&os.ModeSymlink != 0 {
			sourceFilePath, err := os.Readlink(filePath)
			if err != nil {
				return appConfigs, err
			}
			destFilePath = sourceFilePath
		}
		if !path.IsAbs(destFilePath) {
			//base := path.Base(filePath)
			//filepath.Rel()
			absFilePath, err  := filepath.EvalSymlinks(filePath)
			if err != nil {
				fmt.Print(" Error: ", err, " reading file ", filePath, "\n")
			}
			destFilePath = absFilePath
		}
		if len(destFilePath) > 0 {
			content, err := ioutil.ReadFile(destFilePath)
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
		} else {
			appC.Enable = false
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
