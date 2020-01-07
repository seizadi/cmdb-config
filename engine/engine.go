package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/seizadi/cmdb/pkg/pb"
	"os/exec"
	"strings"
	"errors"
)
import "github.com/seizadi/template-engine/util"

// TODO - Check errors from file operations

func CreateValues(req TemplateRequest, providers *pb.ListCloudProvidersResponse) (*pb.Application, error) {
	// TODO - We pick the first one but we should go through the list
	
	ok := false
	var provider *pb.CloudProvider
	for _,provider = range providers.Results {
		if provider.Account == req.ProviderAccount {
			ok = true
			break
		}
	}
	
	if (!ok) {
		return nil, errors.New(fmt.Sprintf("can't find cloud provider with account %s", req.ProviderAccount))
	}
	
	ok = false
	var region *pb.Region
	for _,region = range provider.Regions {
		if region.Name == req.RegionName {
			ok = true
			break
		}
	}
	
	if (!ok) {
		return nil, errors.New(fmt.Sprintf("can't find cloud provider region %s", req.RegionName))
	}
	
	ok = false
	var environment *pb.Environment
	for _,environment = range region.Environments {
		if environment.Name == req.EnviromentName {
			ok = true
			break
		}
	}
	
	if (!ok) {
		return nil, errors.New(fmt.Sprintf("can't find environment %s", req.EnviromentName))
	}
	
	ok = false
	var application *pb.Application
	for _,application = range environment.Applications {
		if application.AppName == req.ApplicationName {
			ok = true
			break
		}
	}
	
	if (!ok) {
		return nil, errors.New(fmt.Sprintf("can't find application %s", req.ApplicationName))
	}

	versionTag := application.VersionTag

	err := util.PutText(versionTag.Version, "tmp/release_name.txt")
	if err != nil {
		return nil,err
	}
	
	err = util.PutMap(Application{Application: application},"tmp/application.yaml")
	if err != nil {
		return nil,err
	}
	
	containers := application.Containers
	err = util.PutMap(Containers{Containers: containers},"tmp/containers.yaml")
	if err != nil {
		return nil,err
	}
	
	manifest := application.Manifest
	var manifestValues ManifestValues
	err = json.Unmarshal([]byte(manifest.Values.Value), &manifestValues)
	if err != nil {
		return nil,err
	}
	
	var manifestServices interface{}
	err = json.Unmarshal([]byte(manifest.Services.Value), &manifestServices)
	if err != nil {
		return nil,err
	}
	
	var manifestIngress interface{}
	err = json.Unmarshal([]byte(manifest.Ingress.Value), &manifestIngress)
	if err != nil {
		return nil,err
	}
	
	var resources interface{}
	err = json.Unmarshal([]byte(manifest.Resources.Value), &resources)
	if err != nil {
		return nil,err
	}
	
	var nodeSelector interface{}
	err = json.Unmarshal([]byte(manifest.NodeSelector.Value), &nodeSelector)
	if err != nil {
		return nil,err
	}
	
	var tolerations interface{}
	err = json.Unmarshal([]byte(manifest.Tolerations.Value), &tolerations)
	if err != nil {
		return nil,err
	}
	
	var affinity interface{}
	err = json.Unmarshal([]byte(manifest.Affinity.Value), &affinity)
	if err != nil {
		return nil,err
	}
	err = util.PutMap(Manifest{
		Manifest: manifest,
		ManifestValues: manifestValues.Values,
		Services: manifestServices,
		Ingress: manifestIngress,
		Resources: resources,
		NodeSelector: nodeSelector,
		Tolerations: tolerations,
		Affinity: affinity,
	},"tmp/manifest.yaml")
	if err != nil {
		return nil,err
	}

	
	vault := manifest.Vault
	err = util.PutMap(Vault{Vault: vault},"tmp/vault.yaml")
	if err != nil {
		return nil,err
	}
	
	secrets := vault.Secrets
	err = util.PutMap(Secrets{Secrets: secrets},"tmp/secrets.yaml")
	if err != nil {
		return nil,err
	}
	
	namespace := strings.ToLower(application.Name + "-" + environment.Name + "-" + (environment.Code).String())
	namespace = strings.Replace(namespace, " ", "-", -1)
	err = util.PutMap(NameSpace{NameSpace: namespace},"tmp/name_space.yaml")
	if err != nil {
		return nil,err
	}
	
	return application,nil
}

func GetArchiveFile(app *pb.Application) string {
	// TODO - Implement this feature; I deprecated the Repo field for VersionTag
	// return app.VersionTag.Repo + "/" + app.VersionTag.Version + ".zip"
	return ""
}

func ResolveManifest(path string) error {
	var out bytes.Buffer
	helmCmd := "#!/bin/bash\n" +
		"helm template " +
	"-f ./tmp/name_space.yaml " +
	"-f ./tmp/containers.yaml " +
	"-f ./tmp/manifest.yaml " +
	"-f ./tmp/vault.yaml " +
	"-f ./tmp/secrets.yaml " +
	"-f ./" + path + "/cmdb.yaml " +
	"-n $(cat ./tmp/release_name.txt) " +
	"./"+ path
	cmd := exec.Command("echo", helmCmd)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	util.CopyBufferContents(out.Bytes(), "./tmp/run.sh")
	
	out.Reset()
	cmd = exec.Command("bash", "./tmp/run.sh")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return err
	}
	util.CopyBufferContents(out.Bytes(), "./tmp/resolved.yaml")
	
	return nil
}