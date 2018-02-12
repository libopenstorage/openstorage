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
		restClient, err = volumeclient.NewDriverClient(osdAddress, volumeDriver, volume.APIVersion, "")

		Expect(err).ToNot(HaveOccurred())
		volumedriver = volumeclient.VolumeDriver(restClient)
	})

	AfterEach(func() {

	})

	Describe("Volume Create", func() {

		var (
			volumeID         string
			numVolumesBefore int
			numVolumesAfter  int
		)

		BeforeEach(func() {
			var err error
			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			var err error
			err = volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should create a volume successfully", func() {

			By("Creating the volume")
			var err error

			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "vol-osd-sanity-cd",
					VolumeLabels: map[string]string{
						"class": "sanity-test-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    1,
					HaLevel: 1,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if no of volumes present in cluster increases by 1")

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			Expect(err).NotTo(HaveOccurred())

			numVolumesAfter = len(volumes)

			Expect(numVolumesAfter).To(Equal(numVolumesBefore + 1))

			By("Inspecting the created volume")

			inspectVolumes := []string{volumeID}
			volumesList, err := volumedriver.Inspect(inspectVolumes)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumesList).NotTo(BeEmpty())
			Expect(len(volumesList)).Should(BeEquivalentTo(1))
			Expect(volumesList[0].GetId()).Should(BeEquivalentTo(volumeID))
		})
	})

	Describe("Volume Delete ", func() {

		var (
			volumeID         string
			numVolumesBefore int
		)

		BeforeEach(func() {

			var err error
			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
		})

		It("Should delete a volume successfully", func() {

			By("Creating the volume")
			var err error

			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-delete",
					VolumeLabels: map[string]string{
						"class": "cd-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    1,
					HaLevel: 1,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Deleting the volume")

			err = volumedriver.Delete(volumeID)
			Expect(err).To(Not(HaveOccurred()))

		})

		It("Should fail to delete non-existing volume", func() {
			var err error

			By("Trying to delete a volume that doesn't exist")

			err = volumedriver.Delete("id-doesnt-exist")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Volume Enumerate", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeIDs        []string
			volumesToCreate  = 3
		)

		BeforeEach(func() {
			var err error
			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {

			for _, id := range volumeIDs {
				err := volumedriver.Delete(id)
				Expect(err).ToNot(HaveOccurred())
			}

		})

		It("Should Enumerate all volumes ", func() {

			By("Creating a list of volumes")

			for i := 0; i < volumesToCreate; i++ {

				vr := &api.VolumeCreateRequest{
					Locator: &api.VolumeLocator{
						Name: "classA-enumerate-" + strconv.Itoa(i),
						VolumeLabels: map[string]string{
							"class": "Class-A",
						},
					},
					Source: &api.Source{},
					Spec: &api.VolumeSpec{
						HaLevel: 1,
					},
				}

				volumeID, err := volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())
				Expect(err).NotTo(HaveOccurred())
				Expect(volumeID).ToNot(BeNil())

				volumeIDs = append(volumeIDs, volumeID)
			}

			for i := 0; i < volumesToCreate; i++ {

				vr := &api.VolumeCreateRequest{
					Locator: &api.VolumeLocator{
						Name: "classB-enumerate-" + strconv.Itoa(i),
						VolumeLabels: map[string]string{
							"class": "Class-B",
						},
					},
					Source: &api.Source{},
					Spec: &api.VolumeSpec{
						HaLevel: 1,
					},
				}

				volumeID, err := volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())
				Expect(err).NotTo(HaveOccurred())
				Expect(volumeID).ToNot(BeNil())

				volumeIDs = append(volumeIDs, volumeID)
			}

			By("Enumerating Class A volumes")

			volumes, err := volumedriver.Enumerate(
				&api.VolumeLocator{},
				map[string]string{
					"class": "Class-A",
				})

			Expect(err).NotTo(HaveOccurred())
			//Expect(len(volumes)).To(BeEquivalentTo(3))

			volumes, err = volumedriver.Enumerate(&api.VolumeLocator{}, map[string]string{
				"class": "Class-A",
			})
			numVolumesAfter = len(volumes)

			Expect(err).NotTo(HaveOccurred())
			//Expect(len(volumes)).To(BeEquivalentTo(0))
			//Expect(len(volumes)).To(BeEquivalentTo(numVolumesBefore + 6))

		})
	})

	Describe("Volume Mount", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
		)

		BeforeEach(func() {
			var err error
			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			var err error

			err = volumedriver.Unmount(volumeID, "/mnt", nil)
			Expect(err).ToNot(HaveOccurred())

			err = volumedriver.Detach(volumeID, nil)
			Expect(err).ToNot(HaveOccurred())

			err = volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should mount successfully", func() {

			By("Creating the volume")

			var err error

			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-mount",
					VolumeLabels: map[string]string{
						"class": "mount-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    1,
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			numVolumesAfter = len(volumes)

			Expect(numVolumesAfter).To(Equal(numVolumesBefore + 1))

			By("Doing a Volume Set to mount the volume")

			// req := &api.VolumeSetRequest{
			// 	Options: map[string]string{},
			// 	Action: &api.VolumeStateAction{
			// 		Attach:    api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			// 		Mount:     api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			// 		MountPath: "/mnt/testmountvolume",
			// 	},
			// 	Locator: &api.VolumeLocator{Name: vr.GetLocator().GetName()},
			// 	Spec:    &api.VolumeSpec{Size: vr.GetSpec().GetSize()},
			// }

			_, err = volumedriver.Attach(volumeID, nil)
			Expect(err).NotTo(HaveOccurred())
			err = volumedriver.Mount(volumeID, "/mnt", nil)

			//err = volumedriver.Set(volumeID, req.GetLocator(), req.GetSpec())

			Expect(err).NotTo(HaveOccurred())

			By("Inspecting the volume and checking attached_on field ")

			volumes, err = volumedriver.Inspect([]string{volumeID})

			Expect(err).NotTo(HaveOccurred())
			//Expect(volumes[0].GetAttachedOn()).To(BeEquivalentTo("/mnt"))

		})
	})

	Describe("Volume Attach", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
		)

		BeforeEach(func() {
			var err error
			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			var err error

			err = volumedriver.Detach(volumeID, nil)
			Expect(err).ToNot(HaveOccurred())

			err = volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should attach successfully", func() {

			By("Creating the volume")

			var err error

			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-attach",
					VolumeLabels: map[string]string{
						"class": "attach-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    1,
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			numVolumesAfter = len(volumes)

			Expect(numVolumesAfter).To(Equal(numVolumesBefore + 1))

			By("Doing a Volume Set to attach the volume to a node")

			// req := &api.VolumeSetRequest{
			// 	Options: map[string]string{},
			// 	Action: &api.VolumeStateAction{
			// 		Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			// 		Mount:  api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF,
			// 	},
			// 	Locator: &api.VolumeLocator{Name: vr.GetLocator().GetName()},
			// 	Spec:    &api.VolumeSpec{Size: vr.GetSpec().GetSize()},
			// }

			//err = volumedriver.Set(volumeID, req.GetLocator(), req.GetSpec())

			str, err := volumedriver.Attach(volumeID, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(str).NotTo(BeNil())

			By("Inspecting the volume and checking attached_on field ")

			volumes, err = volumedriver.Inspect([]string{volumeID})

			Expect(err).NotTo(HaveOccurred())
			//Expect(volumes[0].GetAttachedOn()).To(BeEquivalentTo(""))

		})
	})

	Describe("Volume Update [Has issue]", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
		)

		BeforeEach(func() {
			var err error
			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			var err error
			err = volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should update successfully", func() {

			By("Creating the volume")

			var err error

			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-attach",
					VolumeLabels: map[string]string{
						"class": "attach-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    1,
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			numVolumesAfter = len(volumes)

			Expect(numVolumesAfter).To(Equal(numVolumesBefore + 1))

			By("Updating the volume spec. ")
			newSize := 5

			set := &api.VolumeSetRequest{
				Locator: vr.GetLocator(),
				Spec: &api.VolumeSpec{
					Size:   uint64(newSize),
					Shared: true,
				},
			}

			err = volumedriver.Set(volumeID, set.GetLocator(), set.GetSpec())
			Expect(err).NotTo(HaveOccurred())

			By("Inspecting the volume for new updates")

			volumes, err = volumedriver.Inspect([]string{volumeID})
			Expect(err).NotTo(HaveOccurred())

			Expect(volumes[0].GetSpec().GetSize()).To(BeEquivalentTo(newSize))
			Expect(volumes[0].GetSpec().GetShared()).To(BeTrue())

		})
	})
})
