/*
Copyright 2017 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sanity

import (
	//"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/cluster"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster Data [Cluster Tests]", func() {
	var (
		restClient *client.Client
		manager    cluster.Cluster
	)

	BeforeEach(func() {
		var err error
		restClient, err = clusterclient.NewClusterClient(osdAddress, cluster.APIVersion)
		Expect(err).ToNot(HaveOccurred())
		manager = clusterclient.ClusterManager(restClient)
	})

	Describe("Cluster GetGossipState", func() {
		It("should return a gossip information on the nodes", func() {
			clusterState := manager.GetGossipState()
			Expect(clusterState).NotTo(BeNil())
			Expect(clusterState.NodeStatus).NotTo(BeEmpty())

			By("Checking the node values (Not the status)")
			for _, nv := range clusterState.NodeStatus {
				n, err := manager.Inspect(string(nv.Id))
				Expect(err).NotTo(HaveOccurred())
				Expect(nv.GenNumber).To(Equal(n.GenNumber))
				Expect(nv.LastUpdateTs.IsZero()).NotTo(BeTrue())
			}
		})
	})
})
