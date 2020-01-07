package commands

import (
	"fmt"
	"github.com/seizadi/cmdb/client"
	"github.com/seizadi/template-engine/engine"
	"github.com/seizadi/template-engine/util"
	"log"
	"strings"
	
	"github.com/codegangsta/cli"
)

const srcFile = "./tmp/repo.zip"
const dstDir = "./tmp"

func createManifest(c *cli.Context) {
	
	req := engine.TemplateRequest{
		CmdbHost: c.GlobalString("cmdb"),
		CmdbApiKey: c.GlobalString("apikey"),
		ProviderAccount: c.String("cloud"),
		RegionName: c.String("region"),
		EnviromentName: c.String("environment"),
		ApplicationName: c.String("application"),
		Path: c.String("path"),
	}
	
	h, err := client.NewCmdbClient(req.CmdbHost, req.CmdbApiKey)
	if err != nil {
		return
	}
	
	resp, err := h.GetCloudProviders()
	
	if err != nil {
		fmt.Printf("Error getting Cloud Providers %s\n", err)
		return
	}
	
	application, err := engine.CreateValues(req, resp)
	if err != nil {
		log.Fatalf("Failed to create values %s\n", err)
	}
	
	var path string
	
	if len(req.Path) > 0 {
		path = req.Path
	} else {
		archieveFileUrl := engine.GetArchiveFile(application)
		
		err = util.DownloadFile(srcFile, archieveFileUrl)
		if err != nil {
			log.Fatalf("Failed to download file %s\n", err)
		}
		files := util.UnzipFiles(srcFile, dstDir)
		
		path = getAppRepo(application.AppName, files)
		if len(path) == 0 {
			fmt.Printf("Did not find repo for app %s\n", application.AppName)
			return
		}
	}
		
	err = engine.ResolveManifest(path)
	if err != nil {
		log.Fatalf("Failed to resolve manifest %s\n", err)
	}
	return
}

func getAppRepo(app string, files []string) string {
	for _, f := range files {
		if strings.HasSuffix(f, "/repo/"+ app) {
			fmt.Printf("found file %s\n", f)
			return f
		}
	}
	
	return ""
}