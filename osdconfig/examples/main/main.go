package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/pkg/errors"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

func main() {
	//logrus.SetOutput(ioutil.Discard)
	logrus.SetLevel(logrus.ErrorLevel)
	kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	ctx := context.Background()
	ctx1, cancel1 := context.WithCancel(ctx)
	ctx2, cancel2 := context.WithCancel(ctx1)

	// create a new config manager
	manager, err := osdconfig.NewManager(ctx1, kv)
	if err != nil {
		logrus.Fatal(err)
	}
	defer manager.Close()

	x := make([]int, 0, 0)
	var mu sync.Mutex
	// create a function literal that does something with cluster config
	f := func(config *osdconfig.ClusterConfig) error {
		if config == nil {
			return errors.New("input is nil")
		} else {
			mu.Lock()
			x = append(x, len(x)+1)
			mu.Unlock()
		}
		return nil
	}

	// create a function literal that does something with cluster config
	f2 := func(config *osdconfig.NodeConfig) error {
		if config == nil {
			return errors.New("input is nil")
		} else {
			mu.Lock()
			x = append(x, len(x)+1)
			mu.Unlock()
		}
		return nil
	}

	// register the functional literal to watch cluster config changes
	if err := manager.WatchCluster("clstf", f); err != nil {
		logrus.Fatal(err)
	}
	if err := manager.WatchNode("nodef", f2); err != nil {
		logrus.Fatal(err)
	}

	// update kvdb with config changes
	// it will trigger all the registered callbacks that are listening for such change
	conf := new(osdconfig.ClusterConfig)
	nodeConf := new(osdconfig.NodeConfig)
	t := time.Now()
	go func(ctx context.Context, conf *osdconfig.ClusterConfig) {
		i := 0
		for {
			select {
			case <-ctx.Done():
				logrus.Info("Context cancellation received, exiting")
				fmt.Println("i = ", i*2)
				return
			default:
				conf.ClusterId = fmt.Sprint(i)
				if err := manager.SetClusterConf(conf); err != nil {
					logrus.Fatal(err)
					return
				}

				nodeConf.NodeId = fmt.Sprint(i % 10)
				if err := manager.SetNodeConf(nodeConf); err != nil {
					logrus.Fatal(err)
					return
				}
				i += 1
			}
		}
	}(ctx2, conf)

	// wait a bit for several cluster updates
	time.Sleep(time.Second)
	cancel2()
	cancel1()

	fmt.Println(len(x))
	fmt.Println("All done", time.Since(t))
}
