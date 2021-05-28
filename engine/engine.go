package engine

import (
	"errors"
	"io/ioutil"
	"os"
)

func ConfigCmdb(host string, apiKey string, repo string, chartsRepo string) error {
	basePath := "./tmp"
	destPath := basePath + "/repo"

	// Get Repo
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

	err = createCmdbLifecycle(host, apiKey, lifecycleConfig, chartsRepo)
	if err != nil {
		return err
	}

	return nil
}
