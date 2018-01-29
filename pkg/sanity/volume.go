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
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"

	"github.com/libopenstorage/openstorage/volume"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volume [Volume Tests]", func() {
	var (
		restClient   *client.Client
		volumedriver volume.VolumeDriver
	)

	BeforeEach(func() {
		var err error
		restClient, err = volumeclient.NewDriverClient(osdAddress, "nfs", volume.APIVersion, "")

		Expect(err).ToNot(HaveOccurred())
		//volumedriver = clusterclient.ClusterManager(restClient)
		volumedriver = volumeclient.VolumeDriver(restClient)
	})

	AfterEach(func() {

	})

	Describe("Volume Create", func() {

		var (
			volumerequest *api.VolumeCreateRequest
			volumeID      string
		)

		BeforeEach(func() {

			var err error

			volumerequest = &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "vol-osd-sanity-cd",
					VolumeLabels: map[string]string{
						"class": "sanity-test-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:   1,
					Format: api.FSType_FS_TYPE_NFS,
				},
			}

			volumeID, err = volumedriver.Create(volumerequest.GetLocator(), volumerequest.GetSource(), volumerequest.GetSpec())
			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())
		})

		AfterEach(func() {
			err := volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Volume should be present in inspect", func() {

			inspectVolumes := []string{volumeID}

			By("only one volume to be returned and having the same name as the created volume in beforeeach.")
			volumesList, err := volumedriver.Inspect(inspectVolumes)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumesList).NotTo(BeEmpty())
			Expect(len(volumesList)).Should(BeEquivalentTo(1))
			Expect(volumesList[0].GetId()).Should(BeEquivalentTo(volumeID))
		})

		It("Volume should be created with correct input data", func() {

			inspectVolumes := []string{volumeID}

			By("inspecting the details of the volume")
			volumesList, err := volumedriver.Inspect(inspectVolumes)

			volume := volumesList[0]

			Expect(err).NotTo(HaveOccurred())
			Expect(volume.GetSpec().GetSize()).To(BeEquivalentTo(1))
			Expect(volume.GetSpec().GetFormat()).To(BeEquivalentTo(api.FSType_FS_TYPE_NFS))

		})

		It("should say volume already exists", func() {
			By("Creating the volume with same name again")

			_, err := volumedriver.Create(volumerequest.GetLocator(), volumerequest.GetSource(), volumerequest.GetSpec())

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("already exists"))

		})
	})

	Describe("Volume Enumerate", func() {

		var (
			volumerequest *api.VolumeCreateRequest
			volumeIDs     []string
		)

		BeforeEach(func() {

			for i := 0; i < 5; i++ {

				volumerequest = &api.VolumeCreateRequest{
					Locator: &api.VolumeLocator{
						Name: "vol-osd-sanity-ei-" + strconv.Itoa(i),
						VolumeLabels: map[string]string{
							"class": "sanity-test-class",
						},
					},
					Source: &api.Source{},
					Spec:   &api.VolumeSpec{},
				}

				volumeID, err := volumedriver.Create(volumerequest.GetLocator(), volumerequest.GetSource(), volumerequest.GetSpec())
				Expect(err).NotTo(HaveOccurred())
				Expect(volumeID).ToNot(BeNil())

				volumeIDs = append(volumeIDs, volumeID)
			}
		})

		AfterEach(func() {
			var err error

			for _, id := range volumeIDs {
				err = volumedriver.Delete(id)
				Expect(err).ToNot(HaveOccurred())
			}

			volumeIDs = nil
		})

		It("Should enumerate all created volumes", func() {

			By("All five volumes to be listed")

			volumesList, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			Expect(err).NotTo(HaveOccurred())

			Expect(volumesList).NotTo(BeEmpty())
			Expect(len(volumesList)).Should(BeEquivalentTo(5))
		})

		It("Should enumeate only one volume", func() {
			By("Passing volume name in Volume Locator")

			locator := &api.VolumeLocator{
				Name: "vol-osd-sanity-ei-2",
			}

			volumesList, err := volumedriver.Enumerate(locator, make(map[string]string))

			Expect(err).NotTo(HaveOccurred())

			Expect(volumesList).NotTo(BeEmpty())
			Expect(len(volumesList)).Should(BeEquivalentTo(1))
			Expect(volumesList[0].GetId()).Should(BeEquivalentTo("vol-osd-sanity-ei-2"))

		})
	})

})
