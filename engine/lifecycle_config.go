package engine

import (
	"encoding/json"
	"errors"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/protoc-gen-gorm/types"
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

	//Create CloudProvider and Region to anchor deployment
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

	err = state.visitLifecycle(&resource.Identifier{}, lifecycle)
	if err != nil {
		return err
	}

	return nil
}

func (s *LifecycleState) visitLifecycle(rootId *resource.Identifier, lifecycle LifecycleConfig) error {
	var parentId *resource.Identifier
	var envFlag bool

	if len(lifecycle.LifecycleConfigs) > 0 {
		envFlag = false
		root, err := getLifeCycle(rootId, lifecycle)
		if err != nil {
			return err
		}

		l, err := createLifecycle(s.h, root)
		if err != nil {
			return err
		}
		parentId = l.Id
	} else {
		envFlag = true
		root, err := getEnvironment(rootId, lifecycle)
		if err != nil {
			return err
		}

		e, err := createEnvironment(s.h, root)
		if err != nil {
			return err
		}
		parentId = e.Id
	}

	for _, a := range lifecycle.AppConfigs {
		app, err := s.findCreateApplication(a.Name)
		if err != nil {
			return err
		}

		if len(a.Chart) > 0 && a.Chart != "dnd" && a.Chart != "dev" {
			chart, err := s.findCreateChart(app, a.Chart, false)
			if err != nil {
				return err
			}

			var appVersion pb.AppVersion
			if envFlag {
				appVersion = pb.AppVersion{Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, EnvironmentId: parentId}
			} else {
				appVersion = pb.AppVersion{Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, LifecycleId: parentId}
			}
			_, err = createAppVersion(s.h, appVersion)
			if err != nil {
				return err
			}
		}

		var appConfig pb.AppConfig
		if envFlag {
			appConfig = pb.AppConfig{Name: a.Name, EnvironmentId: parentId, ApplicationId: app.Id}
		} else {
			appConfig = pb.AppConfig{Name: a.Name, LifecycleId: parentId, ApplicationId: app.Id}
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
			err := s.visitLifecycle(parentId, l)
			if err != nil {
				return err
			}
		}

	} else {
		// Create ApplicationInstances
		err := s.createEnvironmentAppInstances(parentId, lifecycle)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *LifecycleState) createEnvironmentAppInstances(rootId *resource.Identifier, lifecycleConfig LifecycleConfig) error {
	var appConfigs []AppConfig
	lifecyclePath := lifecycleConfig.BuildPath
	files, err := ioutil.ReadDir(lifecyclePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := lifecyclePath + "/" + file.Name()
		if file.IsDir() {
			return errors.New("build path has unexpected directory structure")
		} else {
			ext := path.Ext(file.Name())
			if ext == ".yaml" {
				if file.Name() == "values.yaml" {
					// TODO - Not sure why this code returned error!
					//return errors.New("Build path has unexpected values.yaml file.")
					continue
				}
			}
			appConfigs, err = parseAppFiles(file, filePath, appConfigs)
			if err != nil {
				return err
			}
		}
	}

	for _, a := range appConfigs {
		if len(a.ChartFile) > 0 {
			app, err := s.findCreateApplication(a.Name)
			if err != nil {
				return err
			}

			chart, err := s.findCreateChart(app, a.Chart, false)
			if err != nil {
				return err
			}

			appInstance := pb.ApplicationInstance{
				Name:           lifecycleConfig.Name + "/" + app.Name,
				ChartVersionId: chart.Id,
				ApplicationId:  app.Id,
				EnvironmentId:  rootId,
				ConfigYaml: a.Value,
				Enable: a.Enable,
			}

			_, err = createApplicationInstance(s.h, appInstance)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getLifeCycle(lifecycleId *resource.Identifier, lifecycleConfig LifecycleConfig) (pb.Lifecycle, error) {
	var l pb.Lifecycle
	l.Name = lifecycleConfig.Name
	l.LifecycleId = lifecycleId
	l.ConfigYaml = lifecycleConfig.Value
	return l, nil
}

func getEnvironment(lifecycleId *resource.Identifier, lifecycleConfig LifecycleConfig) (pb.Environment, error) {
	var e pb.Environment
	e.Name = lifecycleConfig.Name
	e.LifecycleId = lifecycleId
	e.ConfigYaml = lifecycleConfig.Value
	return e, nil
}

func (s *LifecycleState) findCreateApplication(name string) (pb.Application, error) {
	for _, app := range s.apps {
		if app.Name == name {
			return app, nil
		}
	}

	app := pb.Application{Name: name}
	retApp, err := createApplication(s.h, app)
	if err != nil {
		return app, err
	}
	
	s.apps = append(s.apps, *retApp)
	return *retApp, nil
}

type BuildArtifacts struct {
	Artifacts []BuildChart
}

type BuildChart struct {
	Name      string
	Reference string
	Type      string
}

func (s *LifecycleState) findCreateChart(app pb.Application, chart string, envFlag bool) (pb.ChartVersion, error) {
	var appChart string
	var version string

	appChart = s.chartsRepo + "/" + app.Name
	chartVersion := pb.ChartVersion{Name: appChart + ":" + version, Repo: appChart, ApplicationId: app.Id}
	chartStore :=  types.JSONValue { Value: "{}" }

	if envFlag == false {
		version = chart
	} else {
		// TODO - Do we need the code in else{} clause?
		var build BuildArtifacts
		err := json.Unmarshal([]byte(chart), &build)
		if err != nil {
			return pb.ChartVersion{}, err
		}

		base := path.Base(build.Artifacts[0].Reference)
		reg := regexp.MustCompile(".tgz")
		split := reg.Split(base, -1)
		version = split[0]
		chartStore =  types.JSONValue { Value: chart }
	}
	
	chartVersion.Version = version
	chartVersion.ChartStore = &chartStore

	for i, c := range s.charts {
		if c.Repo == appChart && c.Version == version {
			if envFlag == false {
				return c, nil
			} else {
				c.ChartStore.Value = chartVersion.ChartStore.Value
				res, err := updateChartVersion(s.h, c)
				if err != nil {
					return c, err
				}
				s.charts[i] = *res
				return *res, nil
			}
		}
	}

	retChart, err := createChartVersion(s.h, chartVersion)
	if err != nil {
		return chartVersion, err
	}
	
	s.charts = append(s.charts, *retChart)
	
	return *retChart, nil
}
