package engine

import (
	"errors"
	"io/ioutil"
	"os"
)

func ConfigCmdb(host string, apiKey string, repo string, chartsRepo string) error {
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

	// Parse Repo and Populate CMDB
	var lifecycle os.FileInfo
	files, err := ioutil.ReadDir(destPath)
	for _, f := range files {
		if f.Name() == "envs" {
			lifecycle = f
			break
		}
	}

	if len(lifecycle.Name()) <= 0 {
		return errors.New("Error lifecycle folder " + destPath + "/envs not found")
	}
	//levels := []int {1,2,3}
	lifecycleConfig, err := parseLifecycle(lifecycle, destPath+"/envs")
	if err != nil {
		return err
	}

	createCmdbLifecycle(host, apiKey, lifecycleConfig, chartsRepo)

	return nil
}
