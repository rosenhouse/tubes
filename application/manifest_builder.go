package application

import (
	"fmt"
	"net"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/lib/director"
	"github.com/rosenhouse/tubes/lib/manifests"
)

type directorManifestGenerator interface {
	Generate(config director.DirectorConfig) (manifests.Manifest, error)
}

type boshIOClient interface {
	LatestRelease(releasePath string) (director.Artifact, error)
	LatestStemcell(stemcellName string) (director.Artifact, error)
}

type credentialsGenerator interface {
	Fill(interface{}) error
}

type ManifestBuilder struct {
	DirectorManifestGenerator directorManifestGenerator
	BoshIOClient              boshIOClient
	CredentialsGenerator      credentialsGenerator
}

func (b *ManifestBuilder) getLatestSoftware() (director.Software, error) {
	config := director.Software{}
	var err error
	config.Stemcell, err = b.BoshIOClient.LatestStemcell("bosh-aws-xen-hvm-ubuntu-trusty-go_agent")
	if err != nil {
		return config, err
	}
	config.BoshAWSCPIRelease, err = b.BoshIOClient.LatestRelease("github.com/cloudfoundry-incubator/bosh-aws-cpi-release")
	if err != nil {
		return config, err
	}
	config.BoshDirectorRelease, err = b.BoshIOClient.LatestRelease("github.com/cloudfoundry/bosh")
	if err != nil {
		return config, err
	}
	return config, nil
}

func (b *ManifestBuilder) getInternalIP(subnetCIDR string) (string, error) {
	ip, _, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return "", err
	}
	ip = ip.To4()
	ip = director.IncrementIP(ip, 6)
	return ip.To4().String(), nil
}

func (b *ManifestBuilder) getAWSNetwork(resources awsclient.BaseStackResources) director.AWSNetwork {
	return director.AWSNetwork{
		AvailabilityZone: resources.AvailabilityZone,
		BOSHSubnetID:     resources.BOSHSubnetID,
		BOSHSubnetCIDR:   resources.BOSHSubnetCIDR,
		ElasticIP:        resources.BOSHElasticIP,
		SecurityGroup:    resources.BOSHSecurityGroup,
	}
}

func (b *ManifestBuilder) Build(stackName string, resources awsclient.BaseStackResources, accessKey, secretKey string) ([]byte, error) {
	if accessKey == "" {
		return nil, fmt.Errorf("missing access key")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("missing secret key")
	}

	config := director.DirectorConfig{}

	var err error
	config.Software, err = b.getLatestSoftware()
	if err != nil {
		return nil, err
	}

	err = b.CredentialsGenerator.Fill(&config.Credentials)
	if err != nil {
		return nil, err
	}

	config.InternalIP, err = b.getInternalIP(resources.BOSHSubnetCIDR)
	if err != nil {
		return nil, err
	}

	config.AWSNetwork = b.getAWSNetwork(resources)
	config.AWSSSHKey.Name = stackName
	config.AWSSSHKey.Path = "./ssh-key"
	config.AWSCredentials = director.AWSCredentials{
		Region:          resources.AWSRegion,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}

	manifest, err := b.DirectorManifestGenerator.Generate(config)
	if err != nil {
		return nil, err
	}

	return []byte(manifest.String()), nil
}
