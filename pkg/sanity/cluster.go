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
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"

	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/volume"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster [Cluster Tests]", func() {
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

	AfterEach(func() {
	})

	Describe("Cluster Status", func() {
		It("should have OK status for the cluster", func() {
			By("Enumerating the nodes")
			cluster, err := manager.Enumerate()
			Expect(err).NotTo(HaveOccurred())

			By("checking status the cluster")
			Expect(cluster.Id).NotTo(BeEmpty())
			Expect(cluster.NodeId).NotTo(BeEmpty())
			Expect(cluster.Nodes).NotTo(BeEmpty())
			Expect(cluster.Status).To(Equal(api.Status_STATUS_OK))
		})

		It("should have OK status for all nodes", func() {
			By("Enumerating the nodes")
			cluster, err := manager.Enumerate()
			Expect(err).NotTo(HaveOccurred())

			By("checking status for each node")
			for _, n := range cluster.Nodes {
				Expect(n.Id).NotTo(BeEmpty())
				Expect(n.Hostname).NotTo(BeEmpty())
				Expect(n.Status).To(Equal(api.Status_STATUS_OK))
				Expect(n.Cpu).To(BeNumerically(">", 0))
				Expect(n.MemTotal).To(BeNumerically(">", 0))
				Expect(n.MemUsed).To(BeNumerically(">", 0))
				Expect(n.MemFree).To(BeNumerically(">", 0))
			}
		})
	})

	Describe("Cluster Inspect", func() {
		It("should have ok inspecting each node", func() {
			By("Enumerating the nodes")
			cluster, err := manager.Enumerate()
			Expect(err).NotTo(HaveOccurred())

			By("checking inspecting node")
			for _, n := range cluster.Nodes {
				node, err := manager.Inspect(n.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(node.Id).To(Equal(n.Id))
				Expect(node.Status).To(Equal(n.Status))
				Expect(node.MgmtIp).To(Equal(n.MgmtIp))
				Expect(node.DataIp).To(Equal(n.DataIp))
				Expect(node.Hostname).To(Equal(n.Hostname))
			}
		})
	})

	Describe("Node Status", func() {
		It("Should have an ok status for the node", func() {

			By("Checking the status of the node")
			nodeStatus, err := manager.NodeStatus()
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeStatus).To(BeEquivalentTo(api.Status_STATUS_OK))
		})
	})

	Describe("Gossip State", func() {
		It("Should get Gossip state", func() {

			By("Querying for gossip state")
			gossipState := manager.GetGossipState()
			Expect(len(gossipState.NodeStatus)).ToNot(BeZero())
			Expect(gossipState).NotTo(BeNil())
		})
	})

	Describe("Enumerate Alerts", func() {

		var (
			err              error
			volumeID         string
			numVolumesBefore int
			volumedriver     volume.VolumeDriver
			restClient       *client.Client
		)

		BeforeEach(func() {

			restClient, err = volumeclient.NewDriverClient(osdAddress, volumeDriver, volume.APIVersion, "")
			Expect(err).ToNot(HaveOccurred())
			volumedriver = volumeclient.VolumeDriver(restClient)

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should Enumerate Alerts for volume created / deleted", func() {
			By("Creating the volume first")

			var size = 3
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "volume-to-enumerate-alerts",
					VolumeLabels: map[string]string{
						"class": "enumerate-alerts-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   1,
					BlockSize: 256,
					Format:    api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())
			Expect(err).NotTo(HaveOccurred())

			By("Checking if no of volumes present in cluster increases by 1")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Deleting the created volume")

			err = volumedriver.Delete(volumeID)
			Expect(err).NotTo(HaveOccurred())

			endTime := time.Now()
			// 30 seconds  before
			startTime := endTime.Add(-30 * time.Second)
			alerts, err := manager.EnumerateAlerts(startTime, endTime, api.ResourceType_RESOURCE_TYPE_VOLUME)
			Expect(err).NotTo(HaveOccurred())

			noOfOccurence := 0
			for _, alert := range alerts.Alert {
				if alert.ResourceId == volumeID {
					noOfOccurence++
				}
			}
			// No of occurence should be 2  [one for create and one for delete]
			Expect(noOfOccurence).To(BeEquivalentTo(2))
		})

		It("Should enumerate alerts for all resource types ", func() {

			By("Enumeraing alerts")

			endTime := time.Now()
			startTime := endTime.Add(-5 * time.Hour)

			for _, v := range api.ResourceType_value {
				alerts, err := manager.EnumerateAlerts(startTime, endTime, api.ResourceType(v))
				Expect(err).NotTo(HaveOccurred())
				Expect(alerts).NotTo(BeNil())
			}
		})
	})

	Describe("Clear and Erase Alerts", func() {

		var (
			startTime time.Time
			endTime   time.Time
		)
		It("Should clear and erase alerts", func() {

			By("Taking a random alertID from volume resource type")

			endTime = time.Now()
			startTime = endTime.Add(-5 * time.Hour)

			alerts, err := manager.EnumerateAlerts(startTime, endTime, api.ResourceType_RESOURCE_TYPE_NODE)
			Expect(err).NotTo(HaveOccurred())
			Expect(alerts.Alert).NotTo(BeEmpty())

			randomVolumeAlertID := alerts.Alert[random(0, len(alerts.Alert))].Id

			By("Clear alerts")
			err = manager.ClearAlert(api.ResourceType_RESOURCE_TYPE_NODE, randomVolumeAlertID)
			Expect(err).NotTo(HaveOccurred())

			By("Enumerating the alerts again and checking if the alert cleared")

			alerts, err = manager.EnumerateAlerts(startTime, endTime, api.ResourceType_RESOURCE_TYPE_NODE)

			Expect(err).NotTo(HaveOccurred())
			Expect(alerts).NotTo(BeNil())

			for _, alert := range alerts.Alert {
				if alert.Id == randomVolumeAlertID {
					Expect(alert.Cleared).To(BeTrue())
					break
				}
			}

			By("Erasing alerts")
			err = manager.EraseAlert(api.ResourceType_RESOURCE_TYPE_NODE, randomVolumeAlertID)
			Expect(err).NotTo(HaveOccurred())

			By("Enumerating the alerts again and checking if the alert cleared")

			alerts, err = manager.EnumerateAlerts(startTime, endTime, api.ResourceType_RESOURCE_TYPE_NODE)
			Expect(err).NotTo(HaveOccurred())
			Expect(alerts).NotTo(BeNil())

			noOfOccurence := 0
			for _, alert := range alerts.Alert {
				if alert.Id == randomVolumeAlertID {
					noOfOccurence++
				}
			}
			// Alert should not present
			Expect(noOfOccurence).To(BeEquivalentTo(0))

		})

	})

	Describe("Cluster Enable Gossip", func() {
		It("Should Enable Gossip ", func() {

			By("enabling updates")
			err := manager.EnableUpdates()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Cluster Disable Gossip", func() {
		It("Should Enable Gossip ", func() {

			By("disabling updates")

			err := manager.DisableUpdates()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
