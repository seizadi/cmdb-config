package engine

import (
	"github.com/seizadi/cmdb/client"
	"github.com/seizadi/cmdb/pkg/pb"
	"io/ioutil"
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
	root, err := getLifeCycle(lifecycle)
	if err != nil {
		return err
	}
	
	l_ret, err := createLifecycle(s.h, root)
	if err != nil {
		return err
	}
	
	for _, a := range lifecycle.AppConfigs {
		app, new, err := findCreateApplication(s.h, a.Name, s.apps)
		if err != nil {
			return err
		}
		
		if new == true {
			s.apps = append(s.apps, app)
		}
		
		chart, new, err := findCreateChart(s.h, app, a.Chart, s.charts, s.chartsRepo)
		if err != nil {
			return err
		}
		
		if new == true {
			s.charts = append(s.charts, chart)
		}
		
		appVersion := pb.AppVersion{ Name: app.Name, ChartVersionId: chart.Id, ApplicationId: app.Id, LifecycleId: l_ret.Id }
		_, err = createAppVersion(s.h, appVersion)
		if err != nil {
			return err
		}
		
		appConfig := pb.AppConfig{ Name: a.Name, LifecycleId: l_ret.Id, ApplicationId: app.Id }
		if len(a.ValueFile) > 0 {
			content, err := ioutil.ReadFile(lifecycle.ValueFile)
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
	
	return nil
}

func getLifeCycle(lifecycle LifecycleConfig) (pb.Lifecycle, error) {
	var l pb.Lifecycle
	l.Name = lifecycle.Name
	if len(lifecycle.ValueFile) > 0 {
		content, err := ioutil.ReadFile(lifecycle.ValueFile)
		if err != nil {
			return l, err
		}
		l.ConfigYaml = string(content)
	}

	return l, nil
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

func findCreateChart(h *client.CmdbClient, app pb.Application, version string, charts []pb.ChartVersion, chartRepo string) (pb.ChartVersion, bool, error) {
	
	appChart := chartRepo + "/" + app.Name
	
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
