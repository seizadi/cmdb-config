package engine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func ConfigCmdb( host string, apiKey string, repo string) error {
	basePath := "./tmp"
	destPath := basePath + "/repo"
	//var files []string
	//
	//err := filepath.Walk(destPath, func(path string, info os.FileInfo, err error) error {
	//	files = append(files, path)
	//	return nil
	//})
	//if err != nil {
	//	return err
	//}
	//for _, file := range files {
	//	d, f := path.Split(file)
	//	fmt.Println(d, f)
	//}
	
	//// Get Repo
	//GitPull(repo, basePath, destPath)
	//
	//h, err := client.NewCmdbClient(host, apiKey)
	//if err != nil {
	//	return err
	//}
	//
	
	
	var region os.FileInfo
	files, err := ioutil.ReadDir(destPath)
	for _, f := range files {
		if (f.Name() == "envs") {
			region = f
			break
		}
	}
	
	if len(region.Name()) <= 0 {
		return errors.New("Error Region folder " + destPath + "/envs not found")
	}
	levels := []int {1,2,3}
	tierConfig, err := parseTier(region, destPath + "/envs", levels)
	if err != nil {
		return err
	}
	fmt.Println(tierConfig.Name)
	
	//// Create CloudProvider and Region to anchore deployment
	//cloud := pb.CloudProvider{Name: "AWS", Account: "141960", Provider: pb.Provider_AWS}
	//c_ret, err := createCloudProvider(h, cloud)
	//if err != nil {
	//	return err
	//}
	//
	//region := pb.Region{Name: "us-east-1", CloudProviderId: c_ret.Id}
	//r_ret, err := createRegion(h, region)
	//if err != nil {
	//	return err
	//}
	//
	//// Parse Repo and Populate CMDB
	//
	//stage := pb.Stage{Name: "Development", Type: pb.StageType_DEV, RegionId: r_ret.Id}
	//s_ret, err := createStage(h, stage)
	//if err != nil {
	//	return err
	//}
	//
	//dev1 := pb.Environment{Name: "dev1", StageId: s_ret.Id}
	//e_ret, err := createEnvironment(h, dev1)
	//if err != nil {
	//	return err
	//}
	//
	//cmdb := pb.Application{Name: "CMDB", StageId: s_ret.Id}
	//cmdb_ret, err := createApplication(h, cmdb)
	//if err != nil {
	//	return err
	//}
	//
	//cmdb_chart := pb.ChartVersion{Name: "CMDB", Repo: "stable", Version: "v0.0.10"}
	//chart_ret, err := createChartVersion(h, cmdb_chart)
	//if err != nil {
	//	return err
	//}
	//
	//cmdb1 := pb.ApplicationInstance{Name: "CMDB Dev1", ApplicationId: cmdb_ret.Id, ChartVersionId: chart_ret.Id, EnvironmentId: e_ret.Id}
	//cmdb1_ret, err := createApplicationInstance(h, cmdb1)
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Println("Got CMDB with Id ", cmdb1_ret.Id)
	//
	//resp, err := h.GetCloudProviders()
	//
	//if err != nil {
	//	fmt.Printf("Error getting Cloud Providers %s\n", err)
	//	return err
	//}
	//
	//if len(resp.Results) > 0 {
	//	fmt.Println("Got CloudProvider")
	//}
	
	return nil
}

