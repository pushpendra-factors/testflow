package serviceetcd

import "go.etcd.io/etcd/clientv3"

const PatternServerPrefix = "/factors/services/pattern/"
const ProjectVersionKey = "/factors/metadata/project_version_key"

func (client *EtcdClient) DiscoverPatternServers() ([]KV, error) {
	return client.getWithPrefixSortedByKey(PatternServerPrefix)
}

func (client *EtcdClient) RegisterPatternServer(IP, Port string, leaseId clientv3.LeaseID) error {
	address := IP + ":" + Port
	key := PatternServerPrefix + address
	err := client.Put(key, address, clientv3.WithLease(leaseId))
	return err
}

func (client *EtcdClient) IsRegistered(IP, Port string) (bool, error) {
	address := IP + ":" + Port
	key := PatternServerPrefix + address
	key, err := client.Get(key, clientv3.WithPrefix())
	if err != nil {
		if err == NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (client *EtcdClient) GetProjectVersion() (string, error) {
	key, err := client.Get(ProjectVersionKey)
	return key, err
}

func (client *EtcdClient) SetProjectVersion(version string) error {
	err := client.Put(ProjectVersionKey, version)
	return err
}
