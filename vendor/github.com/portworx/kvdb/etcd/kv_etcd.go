package etcd

import (
	"bytes"
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
	Name                          = "etcd-kv"
	defHost                       = "http://127.0.0.1:4001"
	defaultRetryCount             = 60
	defaultIntervalBetweenRetries = time.Millisecond * 500
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
		// The time required for a request to fail - 30 sec
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

func (kv *EtcdKV) get(key string, recursive, sort bool) (*kvdb.KVPair, error) {
	var err error
	var result *e.Response
	for i := 0; i < defaultRetryCount; i++ {
		result, err = kv.client.Get(context.Background(), key, &e.GetOptions{
			Recursive: recursive,
			Sort:      sort,
		})
		if err == nil {
			return kv.resultToKv(result), nil
		}

		switch err.(type) {
		case *e.ClusterError:
			fmt.Printf("kvdb get error: %v, retry count: %v\n", err, i)
			time.Sleep(defaultIntervalBetweenRetries)
		default:
			etcdErr := err.(e.Error)
			if etcdErr.Code == e.ErrorCodeKeyNotFound {
				return nil, kvdb.ErrNotFound
			}
			return nil, err
		}
	}
	return nil, err
}

func (kv *EtcdKV) Get(key string) (*kvdb.KVPair, error) {
	key = kv.domain + key
	return kv.get(key, false, false)
}

func (kv *EtcdKV) GetVal(key string, val interface{}) (*kvdb.KVPair, error) {
	kvp, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(kvp.Value, val)
	if err != nil {
		return kvp, kvdb.ErrUnmarshal
	} else {
		return kvp, nil
	}
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

func (kv *EtcdKV) setWithRetry(ctx context.Context, key, value string,
	opts *e.SetOptions) (*kvdb.KVPair, error) {
	var (
		err    error
		i      int
		result *e.Response
	)
	for i = 0; i < defaultRetryCount; i++ {
		result, err = kv.client.Set(ctx, key, value, opts)
		if err == nil {
			return kv.resultToKv(result), nil
		}
		switch err.(type) {
		case *e.ClusterError:
			cerr := err.(*e.ClusterError)
			fmt.Printf("kvdb set error: %v %v, retry count: %v\n",
				err, cerr.Detail(), i)
			time.Sleep(defaultIntervalBetweenRetries)
		default:
			goto out
		}
	}

out:
	// It's possible that update succeeded but the re-update failed.
	if i > 0 && i < defaultRetryCount && err != nil {
		kvp, err := kv.get(key, false, false)
		if err == nil && bytes.Equal(kvp.Value, []byte(value)) {
			if opts.PrevExist == e.PrevNoExist {
				kvp.Action = kvdb.KVCreate
			} else {
				kvp.Action = kvdb.KVSet
			}
			return kvp, nil
		}
	}

	return nil, err
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
	return kv.setWithRetry(
		context.Background(),
		key,
		string(b),
		&e.SetOptions{
			TTL: time.Duration(ttl) * time.Second,
		})
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
	return kv.setWithRetry(context.Background(), key, string(b), &e.SetOptions{
		TTL:       time.Duration(ttl) * time.Second,
		PrevExist: e.PrevNoExist,
	})
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
	return kv.setWithRetry(context.Background(), key, string(b), &e.SetOptions{
		TTL:       time.Duration(ttl) * time.Second,
		PrevExist: e.PrevExist,
	})
}

func (kv *EtcdKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = kv.domain + prefix
	var err error
	for i := 0; i < defaultRetryCount; i++ {
		result, err := kv.client.Get(context.Background(), prefix, &e.GetOptions{
			Recursive: true,
			Sort:      true,
		})
		if err == nil {
			return kv.resultToKvs(result), nil
		}
		switch err.(type) {
		case *e.ClusterError:
			fmt.Printf("kvdb set error: %v, retry count: %v\n", err, i)
			time.Sleep(defaultIntervalBetweenRetries)
		default:
			return nil, err
		}
	}
	return nil, err
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
	var keyString string
	if kvPair != nil {
		keyString = kvPair.Key
	}
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			l.Lock()
			for !l.unlocked {
				kvPair.TTL = ttl
				kvp, err := kv.CompareAndSet(
					kvPair,
					kvdb.KVTTL|kvdb.KVModifiedIndex,
					kvPair.Value,
				)
				if err != nil {
					fmt.Printf(
						"Error refreshing lock for key %v: %v\n",
						keyString, err,
					)
					l.err = err
					l.Unlock()
					return
				}
				kvPair.ModifiedIndex = kvp.ModifiedIndex
				break
			}
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
