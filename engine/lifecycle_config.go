package engine

import (
	"encoding/json"
	"errors"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/seizadi/cmdb/client"
	"github.com/seizadi/cmdb/pkg/pb"
	"io/ioutil"
	"path"
	"regexp"
)

type LifecycleState struct {
	h               *client.CmdbClient // The CMDB Client Handle
	apps            []pb.Application
	charts          []pb.ChartVersion
	chartsRepo      string
	lifecycleConfig LifecycleConfig
}

func newLifecycleState(host string, apiKey string, chartsRepo string) (*LifecycleState, error) {
	state := LifecycleState{}
	
	h, err := client.NewCmdbClient(host, apiKey)
	if err != nil {
		return nil, err
	}
	state.h = h
	state.chartsRepo = chartsRepo
	return &state, nil
}

func createCmdbLifecycle(host string, apiKey string, lifecycle LifecycleConfig, chartsRepo string) error {

	state, err := newLifecycleState(host, apiKey, chartsRepo)
	if err != nil {
		return err
	}

	//Create CloudProvider and Region to anchore deployment
	cloud := pb.CloudProvider{Name: "AWS", Account: "141960", Provider: pb.Provider_AWS}
	cRet, err := createCloudProvider(state.h, cloud)
	if err != nil {
		return err
	}

	region := pb.Region{Name: "us-east-1", CloudProviderId: cRet.Id}
	_, err = createRegion(state.h, region)
	if err != nil {
		return err
	}
	
	err = state.visitLifecycle(lifecycle)
	if err != nil {
		return err
	}

	return nil
}

func (s *LifecycleState)visitLifecycle(lifecycle LifecycleConfig) error {
	var rootId *resource.Identifier
	var envFlag bool
	
	if len(lifecycle.LifecycleConfigs) > 0 {
		envFlag = false
		root, err := getLifeCycle(lifecycle)
		if err != nil {
			return err
		}
		
		l_ret, err := createLifecycle(s.h, root)
		if err != nil {
			return err
		}
		rootId = l_ret.Id
	} else {
		envFlag = true
		root, err := getEnvironment(lifecycle)
		if err != nil {
			return err
		}
		
		e_ret, err := createEnvironment(s.h, root)
		if err != nil {
			return err
		}
		rootId = e_ret.Id
	}
	
	for _, a := range lifecycle.AppConfigs {
		app, new, err := findCreateApplication(s.h, a.Name, s.apps)
		if err != nil {
			return err
		}
		
		if new == true {
			s.apps = append(s.apps, app)
		}
		
		chart, new, err := findCreateChart(s.h, app, a.Chart, s.charts, s.chartsRepo, false)
		if err != nil {
			return err
		}
		
		if new == true {
			s.charts = append(s.charts, chart)
		}
	
		var appVersion pb.AppVersion
		if (envFlag) {
			appVersion = pb.AppVersion{Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, EnvironmentId: rootId}
		} else {
			appVersion = pb.AppVersion{Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, LifecycleId: rootId}
			}
		_, err = createAppVersion(s.h, appVersion)
		if err != nil {
			return err
		}
		
		var appConfig pb.AppConfig
		if (envFlag) {
			appConfig = pb.AppConfig{Name: a.Name, EnvironmentId: rootId, ApplicationId: app.Id}
		} else {
			appConfig = pb.AppConfig{Name: a.Name, LifecycleId: rootId, ApplicationId: app.Id}
		}
		if len(a.ValueFile) > 0 {
			content, err := ioutil.ReadFile(a.ValueFile)
			if err != nil {
				return err
			}
			appConfig.ConfigYaml = string(content)
		}
		
		_, err = createAppConfig(s.h, appConfig)
		if err != nil {
			return err
		}
	}
	
	if len(lifecycle.LifecycleConfigs) > 0 {
		for _, l := range lifecycle.LifecycleConfigs {
			err := s.visitLifecycle(l)
			if err != nil {
				return err
			}
		}
		
	} else {
		// Create ApplicationInstances
		err := s.createEnvironmentAppInstances(rootId, lifecycle)
		if err != nil {
			return err
		}
	}
	
	return nil
}


func (s *LifecycleState)createEnvironmentAppInstances(rootId *resource.Identifier, lifecycleConfig LifecycleConfig) error {
	var appConfigs []AppConfig
	lifecyclePath := lifecycleConfig.BuildPath
	files, err := ioutil.ReadDir(lifecyclePath)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		filePath := lifecyclePath + "/" + file.Name()
		if file.IsDir() {
			return errors.New("Build path has unexpected directory structure.")
		} else {
			ext := path.Ext(file.Name())
			if ext == ".yaml" {
				if file.Name() == "values.yaml" {
					return errors.New("Build path has unexpected values.yaml file.")
				}
			}
			appConfigs, err = parseAppFiles(file, filePath, appConfigs, true)
			if err != nil {
				return err
			}
		}
	}
	
	for _, a := range appConfigs {
		if len(a.ChartFile) > 0 {
			app, new, err := findCreateApplication(s.h, a.Name, s.apps)
			if err != nil {
				return err
			}
			
			if new == true {
				s.apps = append(s.apps, app)
			}
			
			chart, new, err := findCreateChart(s.h, app, a.Chart, s.charts, s.chartsRepo, true)
			if err != nil {
				return err
			}
			
			if new == true {
				s.charts = append(s.charts, chart)
			}
			
			appInstance := pb.ApplicationInstance{Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, EnvironmentId: rootId}
			if len(a.ValueFile) > 0 {
				content, err := ioutil.ReadFile(a.ValueFile)
				if err != nil {
					return err
				}
				appInstance.ConfigYaml = string(content)
			}
			
			_, err = createApplicationInstance(s.h, appInstance)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getLifeCycle(lifecycleConfig LifecycleConfig) (pb.Lifecycle, error) {
	var l pb.Lifecycle
	l.Name = lifecycleConfig.Name
	if len(lifecycleConfig.ValueFile) > 0 {
		content, err := ioutil.ReadFile(lifecycleConfig.ValueFile)
		if err != nil {
			return l, err
		}
		l.ConfigYaml = string(content)
	}

	return l, nil
}

func getEnvironment(lifecycleConfig LifecycleConfig) (pb.Environment, error) {
	var e pb.Environment
	e.Name = lifecycleConfig.Name
	if len(lifecycleConfig.ValueFile) > 0 {
		content, err := ioutil.ReadFile(lifecycleConfig.ValueFile)
		if err != nil {
			return e, err
		}
		e.ConfigYaml = string(content)
	}
	
	return e, nil
}

func findCreateApplication(h *client.CmdbClient, name string, apps []pb.Application) (pb.Application, bool, error) {
	for _, app := range apps {
		if app.Name == name {
			return app, false, nil
		}
	}

	app := pb.Application{Name: name}
	retApp, err := createApplication(h, app)
	if err != nil {
		return app, false, err
	}
	return *retApp, true, nil
}

type BuildArtifacts struct{
	Artifacts []BuildChart
}

type BuildChart struct {
	Name string
	Reference string
	Type string
}

func findCreateChart(h *client.CmdbClient, app pb.Application, chart string, charts []pb.ChartVersion, chartRepo string, envFlag bool) (pb.ChartVersion, bool, error) {
	var appChart string
	var version string
	
	appChart = chartRepo + "/" + app.Name
	
	if envFlag == false {
		version = chart
	} else {
		var build BuildArtifacts
		err := json.Unmarshal([]byte(chart), &build)
		if err != nil {
			return pb.ChartVersion{}, false, err
		}
		
		base := path.Base(build.Artifacts[0].Reference)
		reg := regexp.MustCompile(".tgz")
		split := reg.Split(base, -1)
		version = split[0]
	}
	
	for _, c := range charts {
		if c.Repo == appChart && c.Version == version {
			return c, false, nil
		}
	}
	
	chartVersion := pb.ChartVersion{ Name: appChart + ":" + version, Repo: appChart, Version: version, ApplicationId: app.Id }
	retChart, err := createChartVersion(h, chartVersion)
	if err != nil {
		return chartVersion, false, err
	}
	return *retChart, true, nil
}
