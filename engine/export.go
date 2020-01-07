package engine

import "github.com/seizadi/cmdb/pkg/pb"
type ManifestValues struct {
	Values map[string]string `json:"values"`
}

//type ManifestPort struct {
//	Name       string `json:"name"`
//	Port       int    `json:"port"`
//	TargetPort string `json:"targetPort"`
//	NodePort   int    `json:"nodePort"`
//	Protocol   string `json:"protocol"`
//}
//
//type ManifestService struct {
//	Name        string         `json:"name"`
//	ServiceName string         `json:"serviceName"`
//	Type        string         `json:"type"`
//	Ports       []ManifestPort `json:"ports"`
//}
//
//
//type MainfestIngressTls struct {
//	SecretName string   `json:"secretName"`
//	Hosts      []string `json:"hosts"`
//}
//
//type MainfestIngressRule struct {
//	Host       string `json:"host"`
//	SecretName string `json:"secretName"`
//}
//
//type ManifestIngress struct {
//	Enabled     bool                 `json:"enabled"`
//	Annotations map[string]string    `json:"annotations"`
//	Tls         []MainfestIngressTls `json:"tls"`
//	Hosts       []string             `json:"hosts"`
//	Path        string               `json:"path"`
//}

type NameSpace struct {
	NameSpace string `yaml:"nameSpace"`
}

type Environment struct {
	Environment *pb.Environment
}

type Application struct {
	Application *pb.Application
}

type Vault struct {
	Vault *pb.Vault
}

type Secrets struct {
	Secrets []*pb.Secret
}

type Containers struct {
	Containers []*pb.Container
}

type Manifest struct {
	Manifest *pb.Manifest `yaml:"manifest"`
	ManifestValues map[string]string `yaml:"manifestValues"`
	Services interface{} `yaml:"services"`
	Ingress interface{} `yaml:"ingress"`
	Resources interface{} `yaml:"resources"`
	NodeSelector interface{} `yaml:"nodeSelector"`
	Tolerations interface{} `yaml:"tolerations"`
	Affinity interface{} `yaml:"affinity"`
}
