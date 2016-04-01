package etcd

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	e "github.com/coreos/etcd/client"
	"github.com/portworx/kvdb"
)

const (
	Name    = "etcd-kv"
	defHost = "http://etcd.portworx.com:4001"
)

type EtcdKV struct {
	client e.KeysAPI
	domain string
}

type etcdLock struct {
	done     chan struct{}
	unlocked bool
	err      error
	sync.Mutex
}

func EtcdInit(domain string,
	machines []string,
	options map[string]string,
) (kvdb.Kvdb, error) {

	if len(machines) == 0 {
		machines = []string{defHost}
	}
	cfg := e.Config{
		Endpoints: machines,
		Transport: e.DefaultTransport,
		// The time required for a request to fail - 10 sec
		HeaderTimeoutPerRequest: time.Duration(10) * time.Second,
	}
	c, err := e.New(cfg)
	if err != nil {
		return nil, err
	}
	if domain != "" && !strings.HasSuffix(domain, "/") {
		domain = domain + "/"
	}
	kvapi := e.NewKeysAPI(c)
	kv := &EtcdKV{
		client: kvapi,
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
	kvp.KVDBIndex = result.Index
	return kvp
}

func (kv *EtcdKV) resultToKvs(result *e.Response) kvdb.KVPairs {

	kvs := make([]*kvdb.KVPair, len(result.Node.Nodes))
	for i := range result.Node.Nodes {
		kvs[i] = kv.nodeToKv(result.Node.Nodes[i])
		kvs[i].KVDBIndex = result.Index
	}
	return kvs
}

func (kv *EtcdKV) Get(key string) (*kvdb.KVPair, error) {
	key = kv.domain + key
	result, err := kv.client.Get(context.Background(), key, &e.GetOptions{
		Recursive: false,
		Sort:      false,
	})
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
	ttl uint64,
) (*kvdb.KVPair, error) {

	key = kv.domain + key
	b, err := kv.toBytes(val)
	if err != nil {
		return nil, err
	}
	result, err := kv.client.Set(
		context.Background(),
		key,
		string(b),
		&e.SetOptions{
			TTL: time.Duration(ttl) * time.Second,
		},
	)
	if err != nil {
		return nil, err
	}

	return kv.resultToKv(result), err
}

func (kv *EtcdKV) Create(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	key = kv.domain + key

	b, err := kv.toBytes(val)
	if err != nil {
		return nil, err
	}
	result, err := kv.client.Set(context.Background(), key, string(b), &e.SetOptions{
		TTL:       time.Duration(ttl) * time.Second,
		PrevExist: e.PrevNoExist,
	})
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) Update(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	key = kv.domain + key

	b, err := kv.toBytes(val)
	if err != nil {
		return nil, err
	}
	result, err := kv.client.Set(context.Background(), key, string(b), &e.SetOptions{
		TTL:       time.Duration(ttl) * time.Second,
		PrevExist: e.PrevExist,
	})
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = kv.domain + prefix

	result, err := kv.client.Get(context.Background(), prefix, &e.GetOptions{
		Recursive: true,
		Sort:      true,
	})
	if err != nil {
		return nil, err
	}
	return kv.resultToKvs(result), err
}

func (kv *EtcdKV) Delete(key string) (*kvdb.KVPair, error) {
	key = kv.domain + key

	result, err := kv.client.Delete(context.Background(), key, &e.DeleteOptions{
		Recursive: false,
	})
	if err == nil {
		return kv.resultToKv(result), err
	}
	return nil, err
}

func (kv *EtcdKV) DeleteTree(prefix string) error {
	prefix = kv.domain + prefix

	_, err := kv.client.Delete(context.Background(), prefix, &e.DeleteOptions{
		Recursive: true,
	})
	return err
}

func (kv *EtcdKV) Keys(prefix, key string) ([]string, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *EtcdKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte,
) (*kvdb.KVPair, error) {

	opts := &e.SetOptions{
		PrevValue: string(prevValue),
	}
	if (flags & kvdb.KVModifiedIndex) != 0 {
		opts.PrevIndex = kvp.ModifiedIndex
	}
	if (flags & kvdb.KVTTL) != 0 {
		opts.TTL = time.Duration(kvp.TTL)
	}
	result, err := kv.client.Set(
		context.Background(),
		kv.domain+kvp.Key,
		string(kvp.Value),
		opts,
	)
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
) (*kvdb.KVPair, error) {

	result, err := kv.client.Delete(
		context.Background(),
		kv.domain+kvp.Key,
		&e.DeleteOptions{
			PrevValue: string(kvp.Value),
			PrevIndex: kvp.ModifiedIndex,
		})
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *EtcdKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {

	key = kv.domain + key
	go kv.watchStart(key, false, waitIndex, opaque, cb)
	return nil
}

func (kv *EtcdKV) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {

	prefix = kv.domain + prefix
	go kv.watchStart(prefix, true, waitIndex, opaque, cb)
	return nil
}

func (kv *EtcdKV) Lock(key string, ttl uint64) (*kvdb.KVPair, error) {
	key = kv.domain + key
	duration := time.Second

	count := 0
	kvPair, err := kv.Create(key, "locked", ttl)
	for maxCount := 300; err != nil && count < maxCount; count++ {
		time.Sleep(duration)
		kvPair, err = kv.Create(key, "locked", ttl)
	}
	if err != nil {
		return nil, err
	}
	if count > 10 {
		fmt.Printf("ETCD: spent %v iterations locking %v\n", count, key)
	}
	kvPair.TTL = int64(time.Duration(ttl) * time.Second)
	kvPair.Lock = &etcdLock{done: make(chan struct{})}
	go kv.refreshLock(kvPair)

	return kvPair, err
}

func (kv *EtcdKV) refreshLock(kvPair *kvdb.KVPair) {

	l := kvPair.Lock.(*etcdLock)
	ttl := kvPair.TTL
	refresh := time.NewTicker(time.Duration(kvPair.TTL) / 4)
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			l.Lock()
			if l.unlocked {
				l.Unlock()
				return
			}
			kvPair.TTL = ttl
			kvp, err := kv.CompareAndSet(kvPair, kvdb.KVTTL|kvdb.KVModifiedIndex, kvPair.Value)
			if err != nil {
				l.Unlock()
				fmt.Printf("Error refreshing lock for key %v: %v\n",
					string(kvp.Key), err)
				l.err = err
				return
			}
			kvPair.ModifiedIndex = kvp.ModifiedIndex
			l.Unlock()

		case <-l.done:
			return
		}
	}
}

func (kv *EtcdKV) Unlock(kvp *kvdb.KVPair) error {
	l, ok := kvp.Lock.(*etcdLock)
	if !ok {
		return fmt.Errorf("Invalid lock structure for key %v", string(kvp.Key))
	}
	l.Lock()
	// Don't modify kvp here, CompareAndDelete does that.
	_, err := kv.CompareAndDelete(kvp, kvdb.KVFlags(0))
	if err == nil {
		l.unlocked = true
		l.Unlock()
		l.done <- struct{}{}
		return nil
	}
	l.Unlock()
	return err
}

func (kv *EtcdKV) watchStart(key string,
	recursive bool,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) {
	ctx, cancel := context.WithCancel(context.Background())
	watcher := kv.client.Watcher(key, &e.WatcherOptions{
		AfterIndex: waitIndex,
		Recursive:  recursive,
	})
	isCancelSent := false
	for {
		r, watchErr := watcher.Next(ctx)
		if watchErr != nil && !isCancelSent {
			e, ok := watchErr.(e.Error)
			if ok {
				fmt.Printf("Etcd error code %d Message %s Cause %s Index %ju\n",
					e.Code, e.Message, e.Cause, e.Index)
			} else {
				fmt.Printf("Etcd returned an error : %s", watchErr.Error())
			}
			cb(key, opaque, nil, watchErr)
		} else if watchErr == nil && !isCancelSent {
			err := cb(key, opaque, kv.resultToKv(r), nil)
			if err != nil {
				// Cancel the context
				cancel()
				isCancelSent = true
			}
		} else {
			cb(key, opaque, nil, kvdb.ErrWatchStopped)
			break
		}
	}
}

func (kv *EtcdKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func init() {
	kvdb.Register(Name, EtcdInit)
}
