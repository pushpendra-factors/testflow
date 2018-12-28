package serviceetcd

import (
	"context"
	"errors"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var NotFound error = errors.New("KeyNotFound")

type EtcdClient struct {
	c *clientv3.Client
}

func New(endpoints []string) (*EtcdClient, error) {
	v3Client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 15 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdClient{c: v3Client}, err
}

func (e *EtcdClient) Close() error {
	return e.c.Close()
}

func (client *EtcdClient) Put(key, value string, opts ...clientv3.OpOption) error {
	_, err := client.c.Put(context.TODO(), key, value, opts...)
	return err
}

func (client *EtcdClient) Get(key string, opts ...clientv3.OpOption) (string, error) {
	resp, err := client.c.Get(context.TODO(), key, opts...)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", NotFound
	}
	return string(resp.Kvs[0].Value), nil
}

func (client *EtcdClient) Del(key string, opts ...clientv3.OpOption) (int64, error) {
	resp, err := client.c.Delete(context.TODO(), key, opts...)
	noOfKeyDeleted := resp.Deleted
	return noOfKeyDeleted, err
}

type KV struct {
	Key   string
	Value string
}

func (client *EtcdClient) getWithPrefixSortedByKey(prefix string) ([]KV, error) {
	resp, err := client.c.Get(context.TODO(), prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return []KV{}, err
	}

	if len(resp.Kvs) == 0 {
		return []KV{}, NotFound
	}

	kVs := make([]KV, 0, 0)
	for _, respKV := range resp.Kvs {
		kv := KV{Key: string(respKV.Key), Value: string(respKV.Value)}
		kVs = append(kVs, kv)
	}
	return kVs, nil
}

func (client *EtcdClient) GrantLease(ttlSeconds int64) (*clientv3.LeaseGrantResponse, error) {
	resp, err := client.c.Grant(context.TODO(), ttlSeconds)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (client *EtcdClient) Watch(key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	wch := client.c.Watch(context.TODO(), key, opts...)
	return wch
}

func (client *EtcdClient) KeepAlive(leaseId clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	ch, err := client.c.KeepAlive(context.TODO(), leaseId)
	return ch, err
}
