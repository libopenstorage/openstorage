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
		{verb: "PUT", path: clusterPath("/alerts/{resource}/{id}", cluster.APIVersion), fn: c.clearAlert},
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
	}
}
