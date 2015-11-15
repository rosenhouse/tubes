package director

import (
	"fmt"
	"net"

	. "github.com/rosenhouse/tubes/lib/manifests"
)

type DirectorConfig struct {
	Software       Software
	Credentials    Credentials
	InternalIP     string
	AWSNetwork     AWSNetwork
	AWSCredentials AWSCredentials
	AWSSSHKey      AWSSSHKey
}

type Software struct {
	BoshDirectorRelease Artifact
	BoshAWSCPIRelease   Artifact
	Stemcell            Artifact
}

type Artifact struct {
	URL string
	SHA string
}

type Credentials struct {
	MBus              string
	NATS              string
	Redis             string
	Postgres          string
	Registry          string
	BlobstoreDirector string
	BlobstoreAgent    string
	HM                string
	Admin             string
}

type AWSNetwork struct {
	AvailabilityZone string
	BOSHSubnetCIDR   string
	BOSHSubnetID     string
	ElasticIP        string
	SecurityGroup    string
}

type AWSCredentials struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type AWSSSHKey struct {
	Name string
	Path string
}

var defaultEphemeralDisk = EphemeralDisk{
	Size: 25000,
	Type: "gp2",
}

var defaultDiskPool = DiskPool{
	Name:            "disks",
	DiskSize:        20000,
	CloudProperties: DiskPoolCloudProperties{Type: "gp2"},
}

var defaultInstanceType = "m3.xlarge"

func IncrementIP(ip net.IP, amount byte) net.IP {
	cloned := append([]byte(nil), ip...)
	cloned[3] += amount
	return cloned
}

func convertSubnet(subnetCIDR, subnetID string) (Subnet, error) {
	_, ipnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return Subnet{}, err
	}
	gateway := IncrementIP(ipnet.IP, 1)
	dns := IncrementIP(ipnet.IP, 2)
	return Subnet{
		Range:           ipnet.String(),
		Gateway:         gateway.String(),
		DNS:             []string{dns.String()},
		CloudProperties: SubnetCloudProperties{Subnet: subnetID},
	}, nil
}

type DirectorManifestGenerator struct{}

func (g DirectorManifestGenerator) Generate(d DirectorConfig) (Manifest, error) {

	awsProperties := map[interface{}]interface{}{
		"access_key_id":           d.AWSCredentials.AccessKeyID,
		"secret_access_key":       d.AWSCredentials.SecretAccessKey,
		"default_key_name":        d.AWSSSHKey.Name,
		"default_security_groups": []interface{}{d.AWSNetwork.SecurityGroup},
		"region":                  d.AWSCredentials.Region,
	}

	privateSubnet, err := convertSubnet(d.AWSNetwork.BOSHSubnetCIDR, d.AWSNetwork.BOSHSubnetID)
	if err != nil {
		return Manifest{}, err
	}

	privateNetwork := Network{
		Name:    "private",
		Type:    "manual",
		Subnets: []Subnet{privateSubnet},
	}
	eipNetwork := Network{
		Name: "public",
		Type: "vip",
	}
	networks := []Network{privateNetwork, eipNetwork}

	resourcePools := []ResourcePool{
		{
			Name:    "vms",
			Network: privateNetwork.Name,
			Stemcell: Stemcell{
				URL:  d.Software.Stemcell.URL,
				SHA1: d.Software.Stemcell.SHA,
			},
			CloudProperties: ResourcePoolCloudProperties{
				InstanceType:     defaultInstanceType,
				AvailabilityZone: d.AWSNetwork.AvailabilityZone,
				EphemeralDisk:    defaultEphemeralDisk,
			},
		},
	}

	diskPools := []DiskPool{defaultDiskPool}

	boshRelease := Release{
		Name: "bosh",
		URL:  d.Software.BoshDirectorRelease.URL,
		SHA1: d.Software.BoshDirectorRelease.SHA,
	}
	cpiRelease := Release{
		Name: "bosh-aws-cpi",
		URL:  d.Software.BoshAWSCPIRelease.URL,
		SHA1: d.Software.BoshAWSCPIRelease.SHA,
	}
	releases := []Release{boshRelease, cpiRelease}

	postgresProperties := map[interface{}]interface{}{
		"listen_address": "127.0.0.1",
		"host":           "127.0.0.1",
		"user":           "postgres",
		"password":       d.Credentials.Postgres,
		"database":       "bosh",
		"adapter":        "postgres",
	}

	ntpProperties := []interface{}{
		"0.pool.ntp.org",
		"1.pool.ntp.org",
	}

	job := Job{
		Name:      "bosh",
		Instances: 1,
		Templates: []Template{
			{"nats", boshRelease.Name},
			{"redis", boshRelease.Name},
			{"postgres", boshRelease.Name},
			{"blobstore", boshRelease.Name},
			{"director", boshRelease.Name},
			{"health_monitor", boshRelease.Name},
			{"registry", boshRelease.Name},
			{"aws_cpi", cpiRelease.Name},
		},
		ResourcePool:       resourcePools[0].Name,
		PersistentDiskPool: diskPools[0].Name,
		Networks: []NetworkReference{
			{
				Name:      privateNetwork.Name,
				StaticIPs: []string{d.InternalIP},
				Default:   []string{"dns", "gateway"},
			},
			{
				Name:      eipNetwork.Name,
				StaticIPs: []string{d.AWSNetwork.ElasticIP},
			},
		},
		Properties: map[string]interface{}{
			"nats": map[interface{}]interface{}{
				"address":  "127.0.0.1",
				"user":     "nats",
				"password": d.Credentials.NATS,
			},
			"redis": map[interface{}]interface{}{
				"listen_address": "127.0.0.1",
				"address":        "127.0.0.1",
				"password":       d.Credentials.Redis,
			},
			"postgres": postgresProperties,
			"registry": map[interface{}]interface{}{
				"address": d.InternalIP,
				"host":    d.InternalIP,
				"db":      postgresProperties,
				"http": map[interface{}]interface{}{
					"user":     "admin",
					"password": d.Credentials.Registry,
					"port":     25777,
				},
				"username": "admin",
				"password": d.Credentials.Registry,
				"port":     25777,
			},
			"blobstore": map[interface{}]interface{}{
				"address":  d.InternalIP,
				"port":     25250,
				"provider": "dav",
				"director": map[interface{}]interface{}{
					"user":     "director",
					"password": d.Credentials.BlobstoreDirector,
				},
				"agent": map[interface{}]interface{}{
					"user":     "agent",
					"password": d.Credentials.BlobstoreAgent,
				},
			},
			"director": map[interface{}]interface{}{
				"address":     "127.0.0.1",
				"name":        "my-bosh",
				"db":          postgresProperties,
				"cpi_job":     "aws_cpi",
				"max_threads": 10,
				"user_management": map[interface{}]interface{}{
					"provider": "local",
					"local": map[interface{}]interface{}{
						"users": []interface{}{
							map[interface{}]interface{}{
								"name":     "admin",
								"password": d.Credentials.Admin,
							},
							map[interface{}]interface{}{
								"name":     "hm",
								"password": d.Credentials.HM,
							},
						},
					},
				},
			},
			"hm": map[interface{}]interface{}{
				"director_account": map[interface{}]interface{}{
					"user":     "hm",
					"password": d.Credentials.HM,
				},
				"resurrector_enabled": true,
			},
			"aws": awsProperties,
			"agent": map[interface{}]interface{}{
				"mbus": fmt.Sprintf("nats://nats:%s@%s:4222",
					d.Credentials.NATS, d.InternalIP),
			},
			"ntp": ntpProperties,
		},
	}

	cloudProvider := CloudProvider{
		Template: Template{
			Name:    "aws_cpi",
			Release: cpiRelease.Name,
		},
		SSHTunnel: SSHTunnel{
			Host:       d.AWSNetwork.ElasticIP,
			Port:       22,
			User:       "vcap",
			PrivateKey: d.AWSSSHKey.Path,
		},
		MBus: fmt.Sprintf("https://mbus:%s@%s:6868", d.Credentials.MBus, d.AWSNetwork.ElasticIP),
		Properties: map[string]interface{}{
			"aws": awsProperties,
			"agent": map[interface{}]interface{}{
				"mbus": fmt.Sprintf("https://mbus:%s@%s:6868", d.Credentials.MBus, "0.0.0.0"),
			},
			"blobstore": map[interface{}]interface{}{
				"provider": "local",
				"path":     "/var/vcap/micro_bosh/data/cache",
			},
			"ntp": ntpProperties,
		},
	}

	manifest := Manifest{
		Name:          "bosh",
		Releases:      releases,
		ResourcePools: resourcePools,
		DiskPools:     diskPools,
		Networks:      networks,
		Jobs:          []Job{job},
		CloudProvider: cloudProvider,
	}

	return manifest, nil
}
