package engine

import (
	"github.com/seizadi/cmdb/client"
	"github.com/seizadi/cmdb/pkg/pb"
)

func createCloudProvider(h *client.CmdbClient, cloud pb.CloudProvider) (*pb.CloudProvider, error) {
	var req pb.CreateCloudProviderRequest
	req.Payload = &cloud
	res, err := h.CreateCloudProvider(&req)
	if err != nil {
		return nil, err
	}
	
	return res.Result, nil
}

func createRegion(h *client.CmdbClient, region pb.Region) (*pb.Region, error) {
	var req pb.CreateRegionRequest
	req.Payload = &region
	res, err := h.CreateRegion(&req)
	if err != nil {
		return nil, err
	}
	
	return res.Result, nil
}


func createStage(h *client.CmdbClient, stage pb.Stage) (*pb.Stage, error) {
	var req pb.CreateStageRequest
	req.Payload = &stage
	res, err := h.CreateStage(&req)
	if err != nil {
		return nil, err
	}
	
	return res.Result, nil
}

func createEnvironment(h *client.CmdbClient, environment pb.Environment) (*pb.Environment, error) {
	var req pb.CreateEnvironmentRequest
	req.Payload = &environment
	res, err := h.CreateEnvironment(&req)
	if err != nil {
		return nil, err
	}
	
	return res.Result, nil
}

func createApplication(h *client.CmdbClient, app pb.Application) (*pb.Application, error) {
	var req pb.CreateApplicationRequest
	req.Payload = &app
	res, err := h.CreateApplication(&req)
	if err != nil {
		return nil, err
	}
	
	return res.Result, nil
}

func createChartVersion(h *client.CmdbClient, chart pb.ChartVersion) (*pb.ChartVersion, error) {
	var req pb.CreateChartVersionRequest
	req.Payload = &chart
	res, err := h.CreateChartVersion(&req)
	if err != nil {
		return nil, err
	}
	
	return res.Result, nil
}

func createApplicationInstance(h *client.CmdbClient, app pb.ApplicationInstance) (*pb.ApplicationInstance, error) {
	var req pb.CreateApplicationInstanceRequest
	req.Payload = &app
	res, err := h.CreateApplicationInstance(&req)
	if err != nil {
		return nil, err
	}
	
	return res.Result, nil
}
