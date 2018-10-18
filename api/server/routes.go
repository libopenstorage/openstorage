package server

import (
	client "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/cluster"
)

func (c *clusterApi) Routes() []*Route {
	return []*Route{
		{verb: "GET", path: "/cluster/versions", fn: c.versions},
		{verb: "GET", path: clusterPath("/enumerate", cluster.APIVersion), fn: c.enumerate},
		{verb: "GET", path: clusterPath("/gossipstate", cluster.APIVersion), fn: c.gossipState},
		{verb: "GET", path: clusterPath("/nodestatus", cluster.APIVersion), fn: c.nodeStatus},
		{verb: "GET", path: clusterPath("/nodehealth", cluster.APIVersion), fn: c.nodeHealth},
		{verb: "GET", path: clusterPath("/status", cluster.APIVersion), fn: c.status},
		{verb: "GET", path: clusterPath("/peerstatus", cluster.APIVersion), fn: c.peerStatus},
		{verb: "GET", path: clusterPath("/inspect/{id}", cluster.APIVersion), fn: c.inspect},
		{verb: "DELETE", path: clusterPath("", cluster.APIVersion), fn: c.delete},
		{verb: "DELETE", path: clusterPath("/{id}", cluster.APIVersion), fn: c.delete},
		{verb: "PUT", path: clusterPath("/enablegossip", cluster.APIVersion), fn: c.enableGossip},
		{verb: "PUT", path: clusterPath("/disablegossip", cluster.APIVersion), fn: c.disableGossip},
		{verb: "PUT", path: clusterPath("/shutdown", cluster.APIVersion), fn: c.shutdown},
		{verb: "PUT", path: clusterPath("/shutdown/{id}", cluster.APIVersion), fn: c.shutdown},
		{verb: "GET", path: clusterPath("/alerts/{resource}", cluster.APIVersion), fn: c.enumerateAlerts},
		{verb: "DELETE", path: clusterPath("/alerts/{resource}/{id}", cluster.APIVersion), fn: c.eraseAlert},
		{verb: "GET", path: clusterPath(client.UriCluster, cluster.APIVersion), fn: c.getClusterConf},
		{verb: "GET", path: clusterPath(client.UriNode+"/{id}", cluster.APIVersion), fn: c.getNodeConf},
		{verb: "GET", path: clusterPath(client.UriEnumerate, cluster.APIVersion), fn: c.enumerateConf},
		{verb: "POST", path: clusterPath(client.UriCluster, cluster.APIVersion), fn: c.setClusterConf},
		{verb: "POST", path: clusterPath(client.UriNode, cluster.APIVersion), fn: c.setNodeConf},
		{verb: "DELETE", path: clusterPath(client.UriNode+"/{id}", cluster.APIVersion), fn: c.delNodeConf},
		{verb: "GET", path: clusterPath("/getnodeidfromip/{idip}", cluster.APIVersion), fn: c.getNodeIdFromIp},
		{verb: "GET", path: clusterSecretPath("/verify", cluster.APIVersion), fn: c.secretLoginCheck},
		{verb: "GET", path: clusterSecretPath("", cluster.APIVersion), fn: c.getSecret},
		{verb: "PUT", path: clusterSecretPath("", cluster.APIVersion), fn: c.setSecret},
		{verb: "GET", path: clusterSecretPath("/defaultsecretkey", cluster.APIVersion), fn: c.getDefaultSecretKey},
		{verb: "PUT", path: clusterSecretPath("/defaultsecretkey", cluster.APIVersion), fn: c.setDefaultSecretKey},
		{verb: "POST", path: clusterSecretPath("/login", cluster.APIVersion), fn: c.secretsLogin},
		{verb: "GET", path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyEnumerate},
		{verb: "GET", path: clusterPath(client.SchedPath+"/{name}", cluster.APIVersion), fn: c.schedPolicyGet},
		{verb: "POST", path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyCreate},
		{verb: "PUT", path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyUpdate},
		{verb: "DELETE", path: clusterPath(client.SchedPath+"/{name}", cluster.APIVersion), fn: c.schedPolicyDelete},
		{verb: "GET", path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreInspect},
		{verb: "POST", path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreCreate},
		{verb: "PUT", path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreUpdate},
		{verb: "DELETE", path: clusterPath(client.ObjectStorePath+"/delete", cluster.APIVersion), fn: c.objectStoreDelete},
		{verb: "PUT", path: clusterPath(client.PairPath, cluster.APIVersion), fn: c.createPair},
		{verb: "POST", path: clusterPath(client.PairPath, cluster.APIVersion), fn: c.processPair},
		{verb: "GET", path: clusterPath(client.PairPath, cluster.APIVersion), fn: c.enumeratePairs},
		{verb: "GET", path: clusterPath(client.PairPath+"/{id}", cluster.APIVersion), fn: c.getPair},
		{verb: "PUT", path: clusterPath(client.PairPath+"/{id}", cluster.APIVersion), fn: c.refreshPair},
		{verb: "DELETE", path: clusterPath(client.PairPath+"/{id}", cluster.APIVersion), fn: c.deletePair},
		{verb: "GET", path: clusterPath(client.PairTokenPath, cluster.APIVersion), fn: c.getPairToken},
	}
}
