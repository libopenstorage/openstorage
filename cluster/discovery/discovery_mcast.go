package discovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"go.pedge.io/dlog"
)

const (
	// DefaultDiscoSrvAddr is a default MultiCast address for multicast discovery type
	DefaultDiscoSrvAddr = "224.0.0.1:9010"
	maxDatagramSize     = 256
)

type discoveryMcast struct {
	sync.Mutex    // embedded
	ClusterInfo   // embedded
	wcb           WatchClusterCB
	clusterName   string
	mcAddr        string
	restAddr      string
	port          string
	restLsnr      net.Listener
	mcLsnr        *net.UDPConn
	MaxUDPPackets int
}

// NewDiscoveryMcast creates a new instance of Multicast kvdb discovery
func NewDiscoveryMcast(clusterName, mcAddr, restAddr string) (mcClust Cluster, err error) {
	if "" == mcAddr {
		mcAddr = DefaultDiscoSrvAddr
	}

	var port string
	if _, port, err = net.SplitHostPort(mcAddr); err != nil {
		// We have to have at least the port
		dlog.WithField("err", err).Errorf("Invalid multicast address provided")
		return
	}

	if "" == restAddr {
		// Rest bind address not provided, let's compute it from Multicast addr
		restAddr = ":" + port
	} else {
		var port0 string
		_, port0, err = net.SplitHostPort(restAddr)
		if err != nil {
			// We have to have at least the port
			dlog.WithField("err", err).Errorf("REST service port not provided")
			return
		} else if port != port0 {
			err = fmt.Errorf("REST service must match multicast port (%s =/= %s)", port0, port)
			return
		}
	}

	mcClust = &discoveryMcast{
		ClusterInfo: ClusterInfo{
			Size:    0,
			Nodes:   make(map[string]NodeEntry),
			Version: 1,
		},
		clusterName:   clusterName,
		mcAddr:        mcAddr,
		restAddr:      restAddr,
		port:          port,
		MaxUDPPackets: 3,
	}
	return
}

func (b *discoveryMcast) AddNode(ne NodeEntry) (*ClusterInfo, error) {
	b.Lock()
	defer b.Unlock()

	if _, exists := b.Nodes[ne.Id]; !exists {
		// We do not exist in the map.
		b.Size++
		b.Version++
		b.Nodes[ne.Id] = ne
	} else if !b.Nodes[ne.Id].Equals(&ne) {
		// bizLogic rule - update only if values were different
		b.Version++
		b.Nodes[ne.Id] = ne
	}

	// create a copy
	ci := *(&b.ClusterInfo)
	if b.wcb != nil {
		b.wcb(&ci, nil)
	}

	return &ci, nil
}

func (b *discoveryMcast) RemoveNode(ne NodeEntry) (*ClusterInfo, error) {
	b.Lock()
	defer b.Unlock()

	_, exists := b.Nodes[ne.Id]
	if !exists {
		return nil, fmt.Errorf("Unable to find node %v in discovery db", ne.Id)
	}
	delete(b.Nodes, ne.Id)
	b.Version++

	// create a copy
	ci := *(&b.ClusterInfo)

	return &ci, nil
}

func (b *discoveryMcast) Enumerate() (*ClusterInfo, error) {
	b.Lock()
	ci := *(&b.ClusterInfo) // Make a copy
	b.Unlock()

	return &ci, nil
}

func (b *discoveryMcast) addAllAndNotifyWatch(stream io.ReadCloser) error {
	var ci ClusterInfo
	if body, err := ioutil.ReadAll(stream); err != nil {
		dlog.WithField("err", err).Warnf("Could not read config-request")
		return err
	} else if err := json.Unmarshal(body, &ci); err != nil {
		dlog.WithField("err", err).Warnf("Could not parse config-request")
		return err
	}

	//// ideally TRACE message
	//logrus.Debugf("addAllAndNotifyWatch() got %+v, have %+v", ci, b.ClusterInfo)

	b.Lock()
	defer b.Unlock()

	stGotNew, stGotUpdates := false, false
	for k, v := range ci.Nodes {
		if _, exists := b.Nodes[k]; !exists {
			dlog.Infof("Discovered node %v", v)
			b.Nodes[k] = v
			stGotNew = true
		} else if ci.Version > b.Version {
			dlog.Infof("Updated node %v", v)
			b.Nodes[k] = v
			stGotUpdates = true
		} else {
			dlog.Debugf("Ignoring update to %s (version too low)", k)
		}
	}

	if ci.Version > b.Version {
		b.Version = ci.Version
	}

	if stGotNew {
		b.Size = len(b.Nodes)
		b.Version++
	}

	// ideally TRACE message
	dlog.Debugf("addAllAndNotifyWatch() -- post additions, my updated state is %v", b.ClusterInfo)

	if b.wcb != nil && (stGotNew || stGotUpdates) {
		ci := *(&b.ClusterInfo)
		b.wcb(&ci, nil)
	}
	return nil
}

// restConfigExchange is a REST-processing function, which will process the passed config
func (b *discoveryMcast) restConfigExchange(w http.ResponseWriter, r *http.Request) {
	dlog.Debugf("Processing REST config update via %s %v", r.Method, r.RequestURI)

	w.Header().Add("Server", "OpenLib-Discovery")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		r.Body.Close()
		return
	}

	// Process input config
	if err := b.addAllAndNotifyWatch(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		r.Body.Close()
		return
	}

	// Send back my updated config
	ci, _ := b.Enumerate()
	reply, err := json.Marshal(ci)
	if err != nil {
		dlog.WithField("err", err).Warnf("Could not encode my nodes")
		return
	}
	_, err = w.Write(reply)
	if err != nil {
		dlog.WithField("err", err).Warnf("Error sending back my nodes")
	} else {
		dlog.Debugf("Sent back my nodes") // string(reply)?
	}
}

func (b *discoveryMcast) mcastConfigExchange(udpAddr *net.UDPAddr) error {
	ci, _ := b.Enumerate()
	payload, err := json.Marshal(ci)
	if err != nil {
		dlog.WithField("err", err).Warnf("Failed serializing config %+v", ci)
		return err
	}

	url := fmt.Sprintf("http://%s:%s", udpAddr.IP.String(), b.port)
	dlog.Debugf("Sending config %v to %s", ci, url)
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		dlog.WithField("err", err).Warnf("Error sendig POST-config request")
	} else if resp.StatusCode != http.StatusOK {
		dlog.Warnf("Got HTTP error [%s] sending config", resp.Status)
	} else {
		dlog.Infof("Configuration sent to %v", url)

		if resp.ContentLength > 0 {
			return b.addAllAndNotifyWatch(resp.Body)
		}
		dlog.Debugf("mcastConfigExchange() my config after update: %v", b.ClusterInfo)
	}
	return nil
}

// pingMulticast transmits a MCAST package, looking for the rest of the cluster.
func (b *discoveryMcast) pingMulticast() error {
	addr, err := net.ResolveUDPAddr("udp", b.mcAddr)
	if err != nil {
		dlog.WithField("err", err).Warnf("Could not resolve UDP for %v", b.mcAddr)
		return err
	}
	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		dlog.WithField("err", err).Warnf("Could not dial multicast on %+v", *addr)
		return err
	}

	dlog.Debugf("Sending multicast ping on %+v", *addr)
	ploadStr := "pwx:" + b.clusterName
	pload := []byte(ploadStr)
	for i := 0; i < b.MaxUDPPackets; i++ {
		if _, err := c.Write(pload); err != nil {
			dlog.WithField("err", err).Warnf("Could not send message")
		} else {
			dlog.Debugf("Multicast message transmitted for cluster %s", ploadStr)
		}
	}
	return nil
}

// serveMcast method starts serving the Multicast Listener running in a loop, listening on multicast requests.
// If request has been received, for the same cluster name, this method will send the latest member-list to the
// caller's REST service.
func (b *discoveryMcast) serveMcast() {
	addr, err := net.ResolveUDPAddr("udp", b.mcAddr)
	if err != nil {
		dlog.WithField("err", err).Errorf("Could not resolve UDP for %v", b.mcAddr)
	}
	dlog.Debugf("Listening Multicast on %+v", *addr)
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		dlog.WithField("err", err).Errorf("Could not set up multicast listener")
	}
	b.mcLsnr = conn
	conn.SetReadBuffer(maxDatagramSize)
	needBytes := []byte("pwx:" + b.clusterName)
	for b.mcLsnr != nil {
		buf := make([]byte, maxDatagramSize)
		n, udpAddr, err := conn.ReadFromUDP(buf)
		if b.mcLsnr == nil {
			// clumsy solution, but looks like there's no better (see https://github.com/golang/go/issues/4373)
			break
		} else if err != nil {
			dlog.WithField("err", err).Warnf("Error reading multicast UDP (sleep & cont..)", err)
			time.Sleep(5 * time.Second)
		} else if n == len(needBytes) && bytes.Equal(buf[:n], needBytes) {
			dlog.Infof("Got config request from client at %v", udpAddr)
			// Send the config back
			b.mcastConfigExchange(udpAddr)
		} else {
			// CAUTION- costly string functions below -- should run only if logger is set to DEBUG, but
			// sadly dlog does not have logrus.GetLevel() equivalent
			dlog.Debugf("Ignoring config request from client at %s :: need '%s', got %v/'%s'",
				udpAddr, string(needBytes), buf[:n], string(buf[:n]))
		}
	}
}

// init will initialize and spawn the IP-service with REST config-service
func (b *discoveryMcast) serveRest() {
	restServer := &http.Server{
		Handler:        http.HandlerFunc(b.restConfigExchange),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 512,
	}

	lsnr, err := net.Listen("tcp", b.restAddr)
	if err != nil {
		dlog.WithField("err", err).Errorf("Could not set up REST-listener")
	}
	b.restLsnr = lsnr

	//http.HandleFunc("/", b.restConfigExchange)
	dlog.Debugf("REST serving on http://%s/%s", lsnr.Addr(), b.clusterName)
	restServer.Serve(lsnr)
}

// Shutdown stops the IP/UDP listeners for multicast discovery.
func (b *discoveryMcast) Shutdown() (err error) {
	// remove watch
	if b.wcb != nil {
		b.wcb = nil
	}
	closing := false
	// stop IP listener
	if b.restLsnr != nil {
		dlog.Debugf("Shutting down REST-Listener")
		if err = b.restLsnr.Close(); err != nil {
			dlog.WithField("err", err).Warnf("Could not shut down the kvdb discovery's REST-listener")
		} else {
			closing = true
			b.restLsnr = nil
		}
	}
	// stop UDP listener
	if b.mcLsnr != nil {
		dlog.Debugf("Shutting down UDP-Listener")
		if err = b.mcLsnr.Close(); err != nil {
			dlog.WithField("err", err).Warnf("Could not shut down the kvdb discovery's UDP-listener")
		} else {
			closing = true
			b.mcLsnr = nil
		}
	}
	if closing {
		// Allow some grace-time if we closed connections
		time.Sleep(2 * time.Second)
	}
	return
}

func (b *discoveryMcast) WatchCluster(wcb WatchClusterCB, lastIndex uint64) error {
	if b.wcb != nil {
		return fmt.Errorf("watch cluster already started")
	}

	b.wcb = wcb
	dlog.Infof("Cluster discovery starting watch at version %d", lastIndex)

	b.Lock()
	b.Version = lastIndex
	b.Unlock()

	// Start the internal services...
	go b.serveRest()
	b.pingMulticast()
	go b.serveMcast()
	return nil
}
