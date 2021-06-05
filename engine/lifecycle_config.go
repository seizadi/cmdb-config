package engine

import (
	"errors"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"io/ioutil"
	"path"

	"github.com/seizadi/cmdb/client"
	"github.com/seizadi/cmdb/pkg/pb"
)

type LifecycleState struct {
	h               *client.CmdbClient // The CMDB Client Handle
	lifecycles	    []pb.Lifecycle
	apps            []pb.Application
	appInstances    []pb.ApplicationInstance // Only for Environments
	appVersions     []pb.AppVersion
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
		s.lifecycles = append(s.lifecycles, *l)
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

		// TODO - We for now ignore 'dnd' and 'dev' chart definitions in lifecycles
		if len(a.Chart) > 0 && a.Chart != "dnd" && a.Chart != "dev" {
			chart, err := s.findCreateChart(app, a.Chart)
			if err != nil {
				return err
			}

			var appVersion pb.AppVersion
			if envFlag {
				appVersion = pb.AppVersion{Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, EnvironmentId: parentId}
			} else {
				appVersion = pb.AppVersion{Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, LifecycleId: parentId}
			}
			appVersionRet, err := createAppVersion(s.h, appVersion)
			if err != nil {
				return err
			}
			s.appVersions = append(s.appVersions, *appVersionRet)
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
		err := s.createEnvironmentAppInstances(rootId, parentId, lifecycle)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *LifecycleState) createEnvironmentAppInstances(lifecycleId *resource.Identifier, envId *resource.Identifier, lifecycleConfig LifecycleConfig) error {
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

	// We create the instances
	appVerMap := s.findAppVersionMap(lifecycleId)
	for _, a := range appConfigs {
		// TODO - This does not handle case where the AppVersion drives the AppInstance creation
		if len(a.ChartFile) > 0 {
			app, err := s.findCreateApplication(a.Name)
			if err != nil {
				return err
			}

			chart, err := s.findCreateChart(app, a.Chart)
			if err != nil {
				return err
			}

			appInstance := pb.ApplicationInstance{
				Name:           lifecycleConfig.Name + "/" + app.Name,
				ChartVersionId: chart.Id,
				ApplicationId:  app.Id,
				EnvironmentId:  envId,
				ConfigYaml:     a.Value,
				Enable:         a.Enable,
			}

			_, err = createApplicationInstance(s.h, appInstance)
			if err != nil {
				return err
			}
		} else {
			appConfig := findAppLifecycleConfig(a.Name, lifecycleConfig)
			if len(a.ValueFile) > 0 || len(appConfig.ValueFile) > 0 {
				appVer := appVerMap[a.Name]
				if appVer.ChartVersionId != nil {
					app, err := s.findCreateApplication(a.Name)
					if err != nil {
						return err
					}

					appInstance := pb.ApplicationInstance{
						Name:           lifecycleConfig.Name + "/" + app.Name,
						ChartVersionId: appVer.ChartVersionId,
						ApplicationId:  app.Id,
						EnvironmentId:  envId,
						ConfigYaml:     a.Value,
						Enable:         true,
					}

					_, err = createApplicationInstance(s.h, appInstance)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func findAppLifecycleConfig(name string, config LifecycleConfig)*AppConfig{
	for i, appConfig := range config.AppConfigs {
		if appConfig.Name == name {
			return &config.AppConfigs[i]
		}
	}
	return nil
}

func (s *LifecycleState) findCreateChart(app pb.Application, version string) (pb.ChartVersion, error) {

	repo := s.chartsRepo + "/" + app.Name
	chartVersion := pb.ChartVersion{
		Name: repo + ":" + version,
		Repo: repo,
		Version: version,
		ApplicationId: app.Id,
	}

	for _, c := range s.charts {
		if c.Name == chartVersion.Name {
			return c, nil
		}
	}

	retChart, err := createChartVersion(s.h, chartVersion)
	if err != nil {
		return chartVersion, err
	}

	s.charts = append(s.charts, *retChart)

	return *retChart, nil
}

func (s *LifecycleState) findLifecycle(lifecycleId *resource.Identifier) *pb.Lifecycle {
	for i, lifecycle := range s.lifecycles {
		if lifecycleId.ResourceId == lifecycle.Id.ResourceId {
			return &s.lifecycles[i]
		}
	}

	return nil
}

func (s *LifecycleState) findLifecycleAppVersion(appId *resource.Identifier, lifecycleId *resource.Identifier) *pb.AppVersion {
	for i, appVersion := range s.appVersions {
		if appVersion.ApplicationId.ResourceId == appId.ResourceId &&
			appVersion.LifecycleId != nil &&
			appVersion.LifecycleId.ResourceId == lifecycleId.ResourceId {
			return &s.appVersions[i]
		}
	}
	return nil
}

func (s *LifecycleState) findAppVersion(id *resource.Identifier, appId *resource.Identifier) pb.AppVersion{
	for id != nil {
		appVersion := s.findLifecycleAppVersion(appId, id)
		if appVersion != nil &&
			appVersion.ChartVersionId != nil {
			return *appVersion
		}
		id = s.findLifecycle(id).LifecycleId
	}
	return pb.AppVersion{}
}

func (s *LifecycleState) findAppVersionMap(lifecycleId *resource.Identifier) map[string]pb.AppVersion {
	// Find all AppVersions that are desired for each Application
	var appVersionMap = make(map[string]pb.AppVersion)
	for _,app := range s.apps {
		appVersionMap[app.Name] = s.findAppVersion(lifecycleId, app.Id)
	}
	return appVersionMap
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

func (s *LifecycleState) findCreateAppInstances(appInstance pb.ApplicationInstance) (pb.ApplicationInstance, error) {
	for _, app := range s.appInstances {
		if app.Name == appInstance.Name {
			return app, nil
		}
	}

	retAppInstance, err := createApplicationInstance(s.h, appInstance)
	if err != nil {
		return appInstance, err
	}

	s.appInstances = append(s.appInstances, *retAppInstance)
	return *retAppInstance, nil
}

type BuildArtifacts struct {
	Artifacts []BuildChart
}

type BuildChart struct {
	Name      string
	Reference string
	Type      string
}

