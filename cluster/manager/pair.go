package manager

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/libopenstorage/openstorage/api"
	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/pkg/errors"
	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
)

const (
	// ClusterPairKey is the key at which info about cluster pairs is stored in kvdb
	ClusterPairKey = "cluster/pair"
	// ClusterPairDefaultKey is the key at which the id for the default pair is stored
	clusterPairDefaultKey = "cluster/pair/default"
)

// CreatePair remote pairs this cluster with a remote cluster.
func (c *ClusterManager) CreatePair(
	request *api.ClusterPairCreateRequest,
) (*api.ClusterPairCreateResponse, error) {
	remoteIp := request.RemoteClusterIp

	for e := c.listeners.Front(); e != nil; e = e.Next() {
		pairMode := e.Value.(cluster.ClusterListener).GetPairMode()
		if pairMode == api.ClusterPairMode_DisasterRecovery &&
			request.Mode == api.ClusterPairMode_Default {
			// If DisasterRecovery mode is set on the listener
			// then override the default mode
			request.Mode = pairMode
		}
	}

	// Pair with remote server
	logrus.Infof("Attempting to pair with cluster at IP %v", remoteIp)
	processRequest := &api.ClusterPairProcessRequest{
		SourceClusterId:    c.Uuid(),
		RemoteClusterToken: request.RemoteClusterToken,
		Mode:               request.Mode,
		CredentialId:       request.CredentialId,
	}

	endpoint := net.JoinHostPort(remoteIp, strconv.FormatUint(uint64(request.RemoteClusterPort), 10))
	clnt, err := clusterclient.NewInsecureTLSAuthClusterClient(endpoint, cluster.APIVersion, request.RemoteClusterToken, "")
	if err != nil {
		return nil, err
	}
	remoteCluster := clusterclient.ClusterManager(clnt)

	// Issue a remote pair request
	resp, err := remoteCluster.ProcessPairRequest(processRequest)
	if err != nil {
		logrus.Warnf("Unable to pair with %v: %v", remoteIp, err)
		return nil, fmt.Errorf("Error from remote cluster: %v", err)
	}

	// Alert all listeners that we are pairing with a cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(cluster.ClusterListener).CreatePair(
			request,
			resp,
		)
		if err != nil {
			logrus.Errorf("Unable to notify %v on a cluster pair event: %v",
				e.Value.(cluster.ClusterListener).String(),
				err,
			)
			return nil, err
		}
	}

	pairInfo := &api.ClusterPairInfo{
		Id:               resp.RemoteClusterId,
		Name:             resp.RemoteClusterName,
		Endpoint:         clnt.BaseURL(),
		CurrentEndpoints: resp.RemoteClusterEndpoints,
		Token:            request.RemoteClusterToken,
		Options:          resp.Options,
		Mode:             request.Mode,
	}

	err = pairCreate(pairInfo, request.SetDefault)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Successfully paired with cluster ID %v", resp.RemoteClusterId)

	response := &api.ClusterPairCreateResponse{
		RemoteClusterId:   pairInfo.Id,
		RemoteClusterName: pairInfo.Name,
	}
	return response, nil
}

// ProcessPairRequest handles a remote cluster's pair request
func (c *ClusterManager) ProcessPairRequest(
	request *api.ClusterPairProcessRequest,
) (*api.ClusterPairProcessResponse, error) {
	if request.SourceClusterId == c.Uuid() {
		return nil, fmt.Errorf("Cannot create cluster pair with self")
	}

	response := &api.ClusterPairProcessResponse{
		RemoteClusterId:   c.Uuid(),
		RemoteClusterName: c.config.ClusterId,
	}

	// Get the token without resetting it
	tokenResp, err := c.GetPairToken(false)
	if err != nil {
		return nil, fmt.Errorf("Error getting Cluster Token: %v", err)
	}
	if tokenResp.Token != request.RemoteClusterToken {
		return nil, fmt.Errorf("Token mismatch during pairing")
	}

	// Alert all listeners that we have received a pair request
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err := e.Value.(cluster.ClusterListener).ProcessPairRequest(
			request,
			response,
		)
		if err != nil {
			logrus.Errorf("Unable to notify %v on a a cluster remote pair request: %v",
				e.Value.(cluster.ClusterListener).String(),
				err,
			)

			return nil, err
		}
	}

	logrus.Infof("Successfully paired with remote cluster %v", request.SourceClusterId)

	return response, nil
}

func (c *ClusterManager) RefreshPair(
	id string,
) error {
	pair, err := pairGet(id)
	if err != nil {
		return err
	}
	processRequest := &api.ClusterPairProcessRequest{
		SourceClusterId:    c.Uuid(),
		RemoteClusterToken: pair.Token,
		CredentialId:       pair.Options[api.OptRemoteCredUUID],
	}

	endpoints := pair.CurrentEndpoints
	endpoints = append(endpoints, pair.Endpoint)
	for _, endpoint := range endpoints {
		clnt, err := clusterclient.NewInsecureTLSAuthClusterClient(endpoint, cluster.APIVersion, pair.Token, "")
		if err != nil {
			logrus.Warnf("Unable to create cluster client for %v: %v", endpoint, err)
			continue
		}
		remoteCluster := clusterclient.ClusterManager(clnt)

		// Issue a remote pair request to get updated info about the cluster
		resp, err := remoteCluster.ProcessPairRequest(processRequest)
		if err != nil {
			logrus.Warnf("Unable to get pair info from %v: %v", endpoint, err)
			continue
		}
		pairInfo := &api.ClusterPairInfo{
			Id:               resp.RemoteClusterId,
			Name:             resp.RemoteClusterName,
			Endpoint:         pair.Endpoint,
			CurrentEndpoints: resp.RemoteClusterEndpoints,
			Token:            pair.Token,
		}
		// Use original options and override with updates ones. This
		// prevents any options we created locally from getting overridden
		pairInfo.Options = pair.Options
		for k, v := range resp.Options {
			pairInfo.Options[k] = v
		}

		return pairUpdate(pairInfo)
	}
	return fmt.Errorf("error updating pair info for %v, all endpoints are unreachable", id)
}

func (c *ClusterManager) DeletePair(
	id string,
) error {
	if err := pairDelete(id); err != nil {
		return err
	}

	// Right now the listeners aren't notified of the delete.
	// Need to add that so that they can stop any operations with that cluster

	logrus.Infof("Successfully deleted pairing with cluster %v", id)
	return nil
}

func (c *ClusterManager) GetPair(
	id string,
) (*api.ClusterPairGetResponse, error) {
	var err error
	if id == "" {
		id, err = getDefaultPairId()
		if err != nil {
			if err == kvdb.ErrNotFound {
				return nil, fmt.Errorf("No default cluster pair found.")
			} else {
				return nil, err
			}
		}
	}
	pair, err := pairGet(id)
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, fmt.Errorf("Cluster pair for id %v not found", id)
		} else {
			return nil, err
		}
	}
	return &api.ClusterPairGetResponse{
		PairInfo: pair,
	}, nil
}

func (c *ClusterManager) EnumeratePairs() (*api.ClusterPairsEnumerateResponse, error) {
	response := &api.ClusterPairsEnumerateResponse{}
	pairs, err := pairList()
	if err != nil {
		return nil, err
	}
	response.Pairs = pairs
	response.DefaultId, err = getDefaultPairId()
	if err != nil {
		logrus.Debugf("Error getting default cluster pair: %v", err)
	}
	return response, nil
}

func (c *ClusterManager) ValidatePair(
	id string,
) error {
	pairResp, err := c.GetPair(id)
	if err != nil {
		return err
	} else if pairResp.PairInfo == nil {
		return fmt.Errorf("Cluster pair for id %v not found", id)
	}

	var lastErr error
	endpoints := pairResp.PairInfo.CurrentEndpoints
	endpoints = append(endpoints, pairResp.PairInfo.Endpoint)
	for _, endpoint := range endpoints {
		clnt, err := clusterclient.NewInsecureTLSAuthClusterClient(
			endpoint,
			cluster.APIVersion,
			pairResp.PairInfo.Token,
			"",
		)
		if err != nil {
			msg := fmt.Sprintf("Unable to create cluster client for %v: %v", endpoint, err)
			logrus.Warn(msg)
			lastErr = fmt.Errorf(msg)
			continue
		}

		resp, err := clusterclient.ClusterManager(clnt).Enumerate()
		if err != nil {
			msg := fmt.Sprintf("Unable to get cluster status from %v: %v", endpoint, err)
			logrus.Warn(msg)
			lastErr = fmt.Errorf(msg)
			continue
		}
		if resp.Status == api.Status_STATUS_OK {
			lastErr = nil
			break
		}
		msg := fmt.Sprintf("Invalid remote cluster status: %v", resp.Status)
		logrus.Warn(msg)
		lastErr = fmt.Errorf(msg)
	}

	if lastErr != nil {
		return lastErr
	}

	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err := e.Value.(cluster.ClusterListener).ValidatePair(
			pairResp.PairInfo,
		)
		if err != nil {
			logrus.Errorf("Unable to validate %v on a cluster validate event: %v",
				e.Value.(cluster.ClusterListener).String(),
				err,
			)
			return fmt.Errorf("Failed to validate cluster pair. %v", err)
		}
	}

	logrus.Infof("Successfully validated pairing with cluster ID: %v",
		pairResp.PairInfo.Id)
	return nil
}

func (c *ClusterManager) GetPairToken(
	reset bool,
) (*api.ClusterPairTokenGetResponse, error) {
	var (
		pairToken   string
		returnError error
	)

	updateCallbackFn := func(db *cluster.ClusterInfo) (bool, error) {
		pairToken = db.PairToken
		// Generate a token if we don't have one or a reset has been requested
		if pairToken == "" || reset {
			token, err := c.generatePairToken()
			if err != nil {
				returnError = errors.Wrap(err, "Failed to generate token")
				return false, nil
			}
			pairToken = fmt.Sprintf("%s", token)
			db.PairToken = pairToken
			return true, nil
		}
		return false, nil
	}

	returnError = updateDB("GetPairToken", c.selfNode.Id, updateCallbackFn)
	return &api.ClusterPairTokenGetResponse{
		Token: pairToken,
	}, returnError
}

func (c *ClusterManager) generatePairToken() (string, error) {
	var token string
	var err error

	if auth.Enabled() {
		token, err = c.systemTokenManager.GetToken(
			&auth.Options{
				Expiration: time.Now().Add(5 * auth.Year).Unix(),
			},
		)
		if err != nil {
			return "", err
		}

	} else {
		randToken := make([]byte, 64)
		rand.Read(randToken)
		token = fmt.Sprintf("%x", randToken)
	}

	return token, nil
}

func pairList() (map[string]*api.ClusterPairInfo, error) {
	kvdb := kvdb.Instance()

	pairs := make(map[string]*api.ClusterPairInfo)
	kv, err := kvdb.Enumerate(ClusterPairKey)
	if err != nil {
		return nil, err
	}

	for _, v := range kv {
		if v.Key == clusterPairDefaultKey {
			continue
		}
		info := &api.ClusterPairInfo{}
		err = json.Unmarshal(v.Value, &info)
		if err != nil {
			return nil, err
		}
		pairs[info.Id] = info
	}

	return pairs, nil
}

func pairCreate(info *api.ClusterPairInfo, setDefault bool) error {
	kv := kvdb.Instance()
	kvp, err := kv.Lock(ClusterPairKey)
	if err != nil {
		return err
	}
	defer kv.Unlock(kvp)

	key := ClusterPairKey + "/" + info.Id
	_, err = kv.Create(key, info, 0)
	if err != nil {
		if err == kvdb.ErrExist {
			kvp, err = kv.Get(key)
			if err != nil {
				return err
			}
			storedInfo := &api.ClusterPairInfo{}
			err = json.Unmarshal(kvp.Value, &storedInfo)
			if err != nil {
				return err
			}
			if info.Token != storedInfo.Token {
				return fmt.Errorf("Invalid token for already paired cluster %v", info.Id)
			}
		} else {
			return err
		}
	}

	defaultId, err := getDefaultPairId()
	// Set this pair as the default if no default is set or it has
	// explicitly been asked
	if setDefault || err == kvdb.ErrNotFound || defaultId == "" {
		err = setDefaultPairId(info.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

// Return the default pair id if set, error if none set
func getDefaultPairId() (string, error) {
	kv := kvdb.Instance()
	kvp, err := kv.Get(clusterPairDefaultKey)
	if err != nil {
		return "", err
	}
	return string(kvp.Value), nil
}

func setDefaultPairId(id string) error {
	kv := kvdb.Instance()
	_, err := kv.Put(clusterPairDefaultKey, id, 0)
	if err != nil {
		return err
	}
	return nil
}

func deleteDefaultPairId() error {
	kv := kvdb.Instance()
	_, err := kv.Delete(clusterPairDefaultKey)
	if err != nil {
		return err
	}
	return nil
}

func pairUpdate(info *api.ClusterPairInfo) error {
	kvdb := kvdb.Instance()
	kvp, err := kvdb.Lock(ClusterPairKey)
	if err != nil {
		return err
	}
	defer kvdb.Unlock(kvp)

	key := ClusterPairKey + "/" + info.Id
	_, err = kvdb.Update(key, info, 0)
	if err != nil {
		return err
	}

	return nil
}

func pairDelete(id string) error {
	kv := kvdb.Instance()
	kvp, err := kv.Lock(ClusterPairKey)
	if err != nil {
		return err
	}
	defer kv.Unlock(kvp)

	defaultId, err := getDefaultPairId()
	if err != kvdb.ErrNotFound && defaultId == id {
		defaultUpdated := false
		// Set one of the other pairs as the default
		pairs, err := pairList()
		if err != nil {
			logrus.Warnf("Error getting clusterpairs, will not update default: %v", err)
		} else {
			for _, pair := range pairs {
				if pair.Id != id {
					err := setDefaultPairId(pair.Id)
					if err != nil {
						logrus.Warnf("Error updating default clusterpair: %v", err)
					} else {
						defaultUpdated = true
						break
					}
				}
			}

		}
		if !defaultUpdated {
			err = deleteDefaultPairId()
			if err != nil {
				return fmt.Errorf("error deleting default pair id")
			}
		}
	}

	key := ClusterPairKey + "/" + id
	_, err = kv.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func pairGet(id string) (*api.ClusterPairInfo, error) {
	kv := kvdb.Instance()
	kvp, err := kv.Lock(ClusterPairKey)
	if err != nil {
		return nil, err
	}
	defer kv.Unlock(kvp)

	key := ClusterPairKey + "/" + id
	kvp, err = kv.Get(key)
	if err != nil {
		return nil, err
	}
	info := &api.ClusterPairInfo{}
	err = json.Unmarshal(kvp.Value, &info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
