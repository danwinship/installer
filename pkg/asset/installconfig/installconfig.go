package installconfig

import (
	"net"
	"os"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	netopv1 "github.com/openshift/cluster-network-operator/pkg/apis/networkoperator/v1"
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/ipnet"
	"github.com/openshift/installer/pkg/types"
)

const (
	installConfigFilename = "install-config.yml"
)

var (
	defaultServiceCIDR      = parseCIDR("10.3.0.0/16")
	defaultClusterCIDR      = "10.2.0.0/16"
	defaultHostSubnetLength = 9 // equivalent to a /23 per node
)

// InstallConfig generates the install-config.yml file.
type InstallConfig struct {
	Config *types.InstallConfig `json:"config"`
	File   *asset.File          `json:"file"`
}

var _ asset.WritableAsset = (*InstallConfig)(nil)

// Dependencies returns all of the dependencies directly needed by an
// InstallConfig asset.
func (a *InstallConfig) Dependencies() []asset.Asset {
	return []asset.Asset{
		&clusterID{},
		&emailAddress{},
		&password{},
		&sshPublicKey{},
		&baseDomain{},
		&clusterName{},
		&pullSecret{},
		&platform{},
	}
}

// Generate generates the install-config.yml file.
func (a *InstallConfig) Generate(parents asset.Parents) error {
	clusterID := &clusterID{}
	emailAddress := &emailAddress{}
	password := &password{}
	sshPublicKey := &sshPublicKey{}
	baseDomain := &baseDomain{}
	clusterName := &clusterName{}
	pullSecret := &pullSecret{}
	platform := &platform{}
	parents.Get(
		clusterID,
		emailAddress,
		password,
		sshPublicKey,
		baseDomain,
		clusterName,
		pullSecret,
		platform,
	)

	a.Config = &types.InstallConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName.ClusterName,
		},
		ClusterID: clusterID.ClusterID,
		Admin: types.Admin{
			Email:    emailAddress.EmailAddress,
			Password: password.Password,
			SSHKey:   sshPublicKey.Key,
		},
		BaseDomain: baseDomain.BaseDomain,
		Networking: types.Networking{
			Type: "OpenshiftSDN",

			ServiceCIDR: ipnet.IPNet{
				IPNet: defaultServiceCIDR,
			},
			ClusterNetworks: []netopv1.ClusterNetwork{
				{
					CIDR:             defaultClusterCIDR,
					HostSubnetLength: uint32(defaultHostSubnetLength),
				},
			},
		},
		PullSecret: pullSecret.PullSecret,
	}

	numberOfMasters := int64(3)
	numberOfWorkers := int64(3)
	switch {
	case platform.AWS != nil:
		a.Config.AWS = platform.AWS
	case platform.OpenStack != nil:
		a.Config.OpenStack = platform.OpenStack
	case platform.Libvirt != nil:
		a.Config.Libvirt = platform.Libvirt
		numberOfMasters = 1
		numberOfWorkers = 1
	default:
		panic("unknown platform type")
	}

	a.Config.Machines = []types.MachinePool{
		{
			Name:     "master",
			Replicas: func(x int64) *int64 { return &x }(numberOfMasters),
		},
		{
			Name:     "worker",
			Replicas: func(x int64) *int64 { return &x }(numberOfWorkers),
		},
	}

	data, err := yaml.Marshal(a.Config)
	if err != nil {
		return errors.Wrap(err, "failed to Marshal InstallConfig")
	}
	a.File = &asset.File{
		Filename: installConfigFilename,
		Data:     data,
	}

	return nil
}

// Name returns the human-friendly name of the asset.
func (a *InstallConfig) Name() string {
	return "Install Config"
}

// Files returns the files generated by the asset.
func (a *InstallConfig) Files() []*asset.File {
	if a.File != nil {
		return []*asset.File{a.File}
	}
	return []*asset.File{}
}

func parseCIDR(s string) net.IPNet {
	_, cidr, _ := net.ParseCIDR(s)
	return *cidr
}

// Load returns the installconfig from disk.
func (a *InstallConfig) Load(f asset.FileFetcher) (found bool, err error) {
	file, err := f.FetchByName(installConfigFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	config := &types.InstallConfig{}
	if err := yaml.Unmarshal(file.Data, config); err != nil {
		return false, errors.Wrapf(err, "failed to unmarshal")
	}

	a.File, a.Config = file, config
	return true, nil
}
