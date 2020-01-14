package engine

import (
	"github.com/seizadi/cmdb-config/util"
	"os"
)

func GitPull(repo string, basePath string, destPath string) (interface{}, error) {

	err := os.MkdirAll(destPath, 0755)
	if err != nil {
		return nil, err
	}

	gitCmd := "#!/bin/bash\n" + "git clone " + repo + " " + destPath
	err = util.CmdExec(gitCmd, basePath+"/fetch.sh")
	if err != nil {
		return nil, err
	}

	return nil, nil
}
