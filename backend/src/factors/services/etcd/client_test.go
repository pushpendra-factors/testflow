package serviceetcd

import (
	U "factors/util"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc/codes"

	"go.etcd.io/etcd/clientv3"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var etcdClient *EtcdClient
var err error

func TestMain(m *testing.M) {
	etcdClient, err = New([]string{"http://localhost:2379"})
	if err != nil {
		log.WithError(err).Fatalln("Failed to initialize etcd client")
	}
	defer etcdClient.Close()
	os.Exit(m.Run())
}

func TestPut(t *testing.T) {

	t.Run("Put", func(t *testing.T) {
		key := U.RandomString(8)
		err := etcdClient.Put(key, U.RandomString(8))
		assert.Nil(t, err)

		//cleanup
		etcdClient.Del(key)
	})

	t.Run("PutWithPrefix", func(t *testing.T) {
		prefix := "/" + U.RandomString(8) + "/"
		key := prefix + U.RandomString(8)
		err := etcdClient.Put(key, U.RandomString(8))
		assert.Nil(t, err)

		//cleanup
		etcdClient.Del(key, clientv3.WithPrefix())
	})

}

func TestGet(t *testing.T) {
	key := U.RandomString(8)
	t.Run("KeyDoesnotExist", func(t *testing.T) {
		// key does not exist, return notFound
		_, err := etcdClient.Get(key)
		assert.Equal(t, NotFound, err)
	})

	t.Run("KeyExists", func(t *testing.T) {
		val := U.RandomString(8)
		err = etcdClient.Put(key, val)
		assert.Nil(t, err)

		rVal, err := etcdClient.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, val, rVal)
	})

	//cleanup
	etcdClient.Del(key)
}

func TestDel(t *testing.T) {
	key := U.RandomString(8)
	val := U.RandomString(8)

	t.Run("DelNotExistingKeys", func(t *testing.T) {
		// No of keys deleted should be 0
		c, err := etcdClient.Del(key)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), c)
	})

	t.Run("DelExistingKeys", func(t *testing.T) {
		// put key
		err = etcdClient.Put(key, val)
		assert.Nil(t, err)

		// No of keys deleted should be 1 now
		c, err := etcdClient.Del(key)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), c)

		_, err = etcdClient.Get(key)
		assert.Equal(t, NotFound, err)
	})

	t.Run("DelExistingKeysByPrefix", func(t *testing.T) {
		prefix := "/" + U.RandomString(8) + "/"
		key := prefix + U.RandomString(8)
		err := etcdClient.Put(key, U.RandomString(8))
		assert.Nil(t, err)

		key2 := prefix + U.RandomString(8)
		err = etcdClient.Put(key2, U.RandomString(8))
		assert.Nil(t, err)

		// No of keys deleted should be 2 now
		c, err := etcdClient.Del(prefix, clientv3.WithPrefix())
		assert.Nil(t, err)
		assert.Equal(t, int64(2), c)

		_, err = etcdClient.Get(prefix, clientv3.WithPrefix())
		assert.Equal(t, NotFound, err)
	})

}

func TestGetWithPrefixSortedByKey(t *testing.T) {
	prefix := "/test/" + U.RandomString(8) + "/"

	k1 := prefix + U.RandomString(8)
	t.Run("KeyDoesnotExist", func(t *testing.T) {
		// key does not exist, return notFound
		_, err := etcdClient.getWithPrefixSortedByKey(k1)
		assert.Equal(t, NotFound, err)
	})

	k2 := prefix + U.RandomString(8)
	val1 := U.RandomString(8)
	val2 := U.RandomString(8)
	t.Run("KeysExistsReturnSortedOrder", func(t *testing.T) {
		err = etcdClient.Put(k1, val1)
		assert.Nil(t, err)

		err = etcdClient.Put(k2, val2)
		assert.Nil(t, err)

		order := []string{k1, k2}
		sort.Strings(order)

		kvs, err := etcdClient.getWithPrefixSortedByKey(prefix)
		assert.Nil(t, err)

		rOrder := []string{}

		for _, kv := range kvs {
			rOrder = append(rOrder, kv.Key)
		}

		assert.Equal(t, order, rOrder)
	})

	// cleanup
	c, err := etcdClient.Del(prefix, clientv3.WithPrefix())
	assert.Nil(t, err)
	assert.Equal(t, int64(2), c)

	_, err = etcdClient.getWithPrefixSortedByKey(prefix)
	assert.Equal(t, NotFound, err)
}

func TestGrantLease(t *testing.T) {

	key := U.RandomString(8)
	val := U.RandomString(8)

	ttl := 1
	lease, err := etcdClient.GrantLease(int64(ttl))
	assert.Nil(t, err)

	err = etcdClient.Put(key, val, clientv3.WithLease(lease.ID))
	assert.Nil(t, err)

	expTTL := 3 * time.Second
	time.Sleep(expTTL)

	// lease & key should be expired by now
	// check by fetching it
	_, err = etcdClient.Get(key)
	assert.Equal(t, NotFound, err)

	// call put on expired lease
	err = etcdClient.Put(key, val, clientv3.WithLease(lease.ID))
	etcderr := err.(rpctypes.EtcdError)
	assert.Equal(t, etcderr.Code(), codes.NotFound)
}
