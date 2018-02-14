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
manager, err := osdconfig.NewManager(context.Background(), kv)
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

# understanding logs
logrus is used to log events. Below is a snapshot of typical log. It shows messages from several `sources`, where
a `callback` is the name of the registered callback, `duration` is the time it took for completion and `msg` is
the short description.

```text
time="2018-02-08T14:05:37-08:00" level=info msg="exec trigger received" execType=osdconfig/clusterConf source="osdconfig/manager/watch"
```
source=`osdconfig/manager/watch` is the watcher for kvdb update and triggers execution of registered callbacks. Currently two
`execTypes` are supported:
* osdconfig/clusterConf: is called when there are updates to cluster config
* osdconfig/nodeConf: is called when there are updates to node config

```text
time="2018-02-08T14:05:37-08:00" level=info msg="spawned successfully" callback=watcher duration=2.459µs source="osdconfig/manager/run" 
```
source=`osdconfig/manager/run` is the scheduler that spawns each registered callbacks. In the log message it indicates that
a particular callback was spawned successfully.

```text
time="2018-02-08T14:05:37-08:00" level=info msg="received data" callback=unknown duration=43.9µs source="osdconfig/manager/run" 
```
In the log message above it indicates that a callback received data. A callback name `unknown` is perfectly valid in this
case since the nature of execution within scheduler makes it impossible to know which callback received the data.

```text
time="2018-02-08T14:05:37-08:00" level=info msg="returned successfully" callback=watcher duration=85.34µs source="osdconfig/manager/run" 
time="2018-02-08T14:05:37-08:00" level=info msg="executed successfully" callback=watcher duration=109.262µs source="osdconfig/manager/printStatus" 
```
And this log message indicates that a particular callback returned successfully and finally captured by the bookkeeping
which prints final exec status

```text
time="2018-02-08T14:05:38-08:00" level=info msg="cancelling contexts" source="osdconfig/manager/close" 
time="2018-02-08T14:05:38-08:00" level=info msg="context cancelled" source="osdconfig/manager/watch" 
time="2018-02-08T14:05:38-08:00" level=info msg="watch stopped" source="osdconfig/manager/watch" 
time="2018-02-08T14:05:39-08:00" level=info msg="releasing memory" source="osdconfig/manager/close" 
time="2018-02-08T14:05:39-08:00" level=info msg="cleanup done" source="osdconfig/manager/close" 
```

Finally, at when the `Close()` is called, contexts get cancelled and watch is stopped.