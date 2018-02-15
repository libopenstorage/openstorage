# osdconfig
osdconfig is a library to work with distributed configuration parameters. It defines an interface and provides an
implementation against KVDB backend.

General purpose of this package is as follows:
* Allow users to register their callback functions on updates to backend
* Allow users to set/get config parameters

# installation
osdconfig is written in go (golang). It can be installed using go get
```bash
$ go get github.com/libopenstorage/openstorage/osdconfig/...
```

# example
Create an instance of kvdb
```go
kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil)
if err != nil {
    logrus.Fatal(err)
}
```

Create an instance of osdconfig manager
```go
manager, err := osdconfig.NewManager(kv)
if err != nil {
	// do something
}
defer manager.Close()
```

Define a function literal that can be registered to watch for changes
```go
f := func(config *osdconfig.Config) error {
	// do something with config (say print)
	fmt.Println(config)
	return nil
}
```

Register this function literal to watch on kvdb changes
```go
if err := manager.WatchCluster("watcher", f); err != nil {
	// do something
}
```

Update cluster config on kvdb
```go
conf := new(osdconfig.ClusterConfig)
conf.ClusterID = "myID"
if err := manager.SetClusterConf(conf); err != nil {
	// do something
	return
}
```

once updates are pushed to backend (kvdb in this implementation), callback function is called