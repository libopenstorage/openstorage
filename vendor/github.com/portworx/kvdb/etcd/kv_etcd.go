package etcd

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	e "github.com/coreos/go-etcd/etcd"
	"github.com/portworx/kvdb"
)

const (
	Name    = "etcd-kv"
	defHost = "http://127.0.0.1:4001"
)

type EtcdKV struct {
	client *e.Client
	domain string
}

func EtcdInit(domain string,
	machines []string,
	options map[string]string) (kvdb.Kvdb, error) {

	if len(machines) == 0 {
		machines = []string{defHost}
	}
	if domain != "" && !strings.HasSuffix(domain, "/") {
		domain = domain + "/"
	}
	kv := &EtcdKV{
		client: e.NewClient(machines),
		domain: domain,
	}
	return kv, nil
}

func (kv *EtcdKV) String() string {
	return Name
}

func (kv *EtcdKV) nodeToKv(node *e.Node) *kvdb.KVPair {

	kvp := &kvdb.KVPair{
		Value:         []byte(node.Value),
		TTL:           node.TTL,
		ModifiedIndex: node.ModifiedIndex,
		CreatedIndex:  node.CreatedIndex,
	}

	// Strip out the leading '/'
	if len(node.Key) != 0 {
		kvp.Key = node.Key[1:]
	} else {
		kvp.Key = node.Key
	}
	kvp.Key = strings.TrimPrefix(kvp.Key, kv.domain)
	return kvp
}

func (kv *EtcdKV) resultToKv(result *e.Response) *kvdb.KVPair {

	kvp := kv.nodeToKv(result.Node)
	switch result.Action {
	case "create":
		kvp.Action = kvdb.KVCreate
	case "set", "update":
		kvp.Action = kvdb.KVSet
	case "delete":
		kvp.Action = kvdb.KVDelete
	case "get":
		kvp.Action = kvdb.KVGet
	default:
		kvp.Action = kvdb.KVUknown
	}
	kvp.KVDBIndex = result.EtcdIndex
	return kvp
}

func (kv *EtcdKV) resultToKvs(result *e.Response) kvdb.KVPairs {

	kvs := make([]*kvdb.KVPair, len(result.Node.Nodes))
	for i := range result.Node.Nodes {
		kvs[i] = kv.nodeToKv(result.Node.Nodes[i])
		kvs[i].KVDBIndex = result.EtcdIndex
	}
	return kvs
}

func (kv *EtcdKV) Get(key string) (*kvdb.KVPair, error) {
	key = kv.domain + key
	result, err := kv.client.Get(key, false, false)
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) GetVal(key string, val interface{}) (*kvdb.KVPair, error) {
	kvp, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(kvp.Value, val)
	return kvp, err
}

func (kv *EtcdKV) toBytes(val interface{}) ([]byte, error) {
	var (
		b   []byte
		err error
	)

	switch val.(type) {
	case string:
		b = []byte(val.(string))
	case []byte:
		b = val.([]byte)
	default:
		b, err = json.Marshal(val)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (kv *EtcdKV) Put(
	key string,
	val interface{},
	ttl uint64) (*kvdb.KVPair, error) {

	key = kv.domain + key
	b, err := kv.toBytes(val)
	if err != nil {
		return nil, err
	}
	result, err := kv.client.Set(key, string(b), ttl)
	if err != nil {
		return nil, err
	}

	return kv.resultToKv(result), err
}

func (kv *EtcdKV) Create(key string, val interface{}, ttl uint64) (*kvdb.KVPair, error) {
	key = kv.domain + key

	b, err := kv.toBytes(val)
	if err != nil {
		return nil, err
	}
	result, err := kv.client.Create(key, string(b), ttl)
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) Update(key string, val interface{}, ttl uint64) (*kvdb.KVPair, error) {
	key = kv.domain + key

	b, err := kv.toBytes(val)
	if err != nil {
		return nil, err
	}
	result, err := kv.client.Update(key, string(b), ttl)
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = kv.domain + prefix

	result, err := kv.client.Get(prefix, true, true)

	if err != nil {
		return nil, err
	}
	return kv.resultToKvs(result), err
}

func (kv *EtcdKV) Delete(key string) (*kvdb.KVPair, error) {
	key = kv.domain + key

	result, err := kv.client.Delete(key, false)
	if err == nil {
		return kv.resultToKv(result), err
	}
	return nil, err
}

func (kv *EtcdKV) DeleteTree(prefix string) error {
	prefix = kv.domain + prefix

	_, err := kv.client.Delete(prefix, true)
	return err
}

func (kv *EtcdKV) Keys(prefix, key string) ([]string, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *EtcdKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte) (*kvdb.KVPair, error) {

	prevIndex := uint64(0)
	if (flags & kvdb.KVModifiedIndex) != 0 {
		prevIndex = kvp.ModifiedIndex
	}
	result, err := kv.client.CompareAndSwap(
		kv.domain+kvp.Key,
		string(kvp.Value),
		0,
		string(prevValue),
		prevIndex)
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags) (*kvdb.KVPair, error) {
	result, err := kv.client.CompareAndDelete(kv.domain+kvp.Key,
		string(kvp.Value),
		kvp.ModifiedIndex)

	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB) error {

	key = kv.domain + key
	go kv.watchStart(key, false, waitIndex, opaque, cb)
	return nil
}

func (kv *EtcdKV) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB) error {

	prefix = kv.domain + prefix
	go kv.watchStart(prefix, true, waitIndex, opaque, cb)
	return nil
}

func (kv *EtcdKV) Lock(key string, ttl uint64) (*kvdb.KVPair, error) {
	key = kv.domain + key

	duration := time.Duration(math.Min(float64(time.Second),
		float64((time.Duration(ttl)*time.Second)/10)))

	result, err := kv.client.Create(key, "locked", ttl)
	for err != nil {
		time.Sleep(duration)
		result, err = kv.client.Create(key, "locked", ttl)
	}

	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) Unlock(kvp *kvdb.KVPair) error {
	// Don't modify kvp here, CompareAndDelete does that.
	_, err := kv.CompareAndDelete(kvp, kvdb.KVFlags(0))
	return err
}

func (kv *EtcdKV) watchReceive(
	key string,
	opaque interface{},
	c chan *e.Response,
	stop chan bool,
	cb kvdb.WatchCB) {

	var err error
	for r, more := <-c; err == nil && more; {
		if more {
			err = cb(key, opaque, kv.resultToKv(r), nil)
			if err == nil {
				r, more = <-c
			}
		}
	}
	stop <- true
	close(stop)
}

func (kv *EtcdKV) watchStart(key string,
	recursive bool,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB) {

	ch := make(chan *e.Response, 10)
	stop := make(chan bool, 1)

	go kv.watchReceive(key, opaque, ch, stop, cb)

	_, err := kv.client.Watch(key, waitIndex, recursive, ch, stop)
	if err != e.ErrWatchStoppedByUser {
		e, ok := err.(e.EtcdError)
		if ok {
			fmt.Printf("Etcd error code %d, message %s cause %s Index %ju\n",
				e.ErrorCode, e.Message, e.Cause, e.Index)
		}
		cb(key, opaque, nil, err)
		fmt.Errorf("Watch returned unexpected error %s\n", err.Error())
	} else {
		cb(key, opaque, nil, kvdb.ErrWatchStopped)
	}
}

func (kv *EtcdKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func init() {
	kvdb.Register(Name, EtcdInit)
}
