package manifests

import "gopkg.in/yaml.v2"

type Manifest struct {
	Name          string         `yaml:"name"`
	Releases      []Release      `yaml:"releases"`
	ResourcePools []ResourcePool `yaml:"resource_pools"`
	DiskPools     []DiskPool     `yaml:"disk_pools"`
	Networks      []Network      `yaml:"networks"`
	Jobs          []Job          `yaml:"jobs"`
	CloudProvider CloudProvider  `yaml:"cloud_provider"`
}

func (m Manifest) String() string {
	s, e := yaml.Marshal(m)
	if e != nil {
		panic(e)
	}
	return string(s)
}

type ResourcePool struct {
	Name            string                      `yaml:"name"`
	Network         string                      `yaml:"network"`
	Stemcell        Stemcell                    `yaml:"stemcell"`
	CloudProperties ResourcePoolCloudProperties `yaml:"cloud_properties"`
}

type Release struct {
	Name string `yaml:"name,omitempty"`
	URL  string `yaml:"url"`
	SHA1 string `yaml:"sha1"`
}

type Stemcell struct {
	URL  string `yaml:"url"`
	SHA1 string `yaml:"sha1"`
}

type ResourcePoolCloudProperties struct {
	InstanceType     string        `yaml:"instance_type"`
	EphemeralDisk    EphemeralDisk `yaml:"ephemeral_disk"`
	AvailabilityZone string        `yaml:"availability_zone"`
}

type EphemeralDisk struct {
	Size int    `yaml:"size"`
	Type string `yaml:"type"`
}

type DiskPool struct {
	Name            string                  `yaml:"name"`
	DiskSize        int                     `yaml:"disk_size"`
	CloudProperties DiskPoolCloudProperties `yaml:"cloud_properties"`
}

type DiskPoolCloudProperties struct {
	Type string `yaml:"type"`
}

type Network struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Subnets []Subnet `yaml:"subnets,omitempty"`
}

type NetworkReference struct {
	Name      string   `yaml:"name"`
	StaticIPs []string `yaml:"static_ips,omitempty"`
	Default   []string `yaml:"default,omitempty"`
}

type Subnet struct {
	Range           string                `yaml:"range"`
	Gateway         string                `yaml:"gateway"`
	DNS             []string              `yaml:"dns"`
	CloudProperties SubnetCloudProperties `yaml:"cloud_properties"`
}

type SubnetCloudProperties struct {
	Subnet string `yaml:"subnet"`
}

type Job struct {
	Name               string                 `yaml:"name"`
	Instances          int                    `yaml:"instances"`
	Templates          []Template             `yaml:"templates"`
	ResourcePool       string                 `yaml:"resource_pool"`
	PersistentDiskPool string                 `yaml:"persistent_disk_pool"`
	Networks           []NetworkReference     `yaml:"networks"`
	Properties         map[string]interface{} `yaml:"properties"`
}

type CloudProvider struct {
	Template   Template               `yaml:"template"`
	SSHTunnel  SSHTunnel              `yaml:"ssh_tunnel"`
	MBus       string                 `yaml:"mbus"`
	Properties map[string]interface{} `yaml:"properties"`
}

type Template struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

type SSHTunnel struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	PrivateKey string `yaml:"private_key"`
}
