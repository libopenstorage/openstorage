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
	"math"
	"strconv"
	"time"

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
			if volumeID != "" {
				err = volumedriver.Delete(volumeID)
				Expect(err).ToNot(HaveOccurred())

				volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
				Expect(err).ToNot(HaveOccurred())
				numVolumesAfter = len(volumes)
			}
		})

		It("Should create a volume successfully", func() {

			By("Creating the volume")
			var err error

			var size = 3
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "vol-osd-sanity-cd",
					VolumeLabels: map[string]string{
						"class": "sanity-test-class",
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
			Expect(volumeID).ToNot(BeNil())
			Expect(volumeID).ToNot(BeEmpty())

			By("Checking if no of volumes present in cluster increases by 1")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)
		})

		It("Should create a shared volume successfully", func() {

			By("Creating the volume")
			var err error
			var size = 3
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "vol-osd-sanity-shared",
					VolumeLabels: map[string]string{
						"class": "sanity-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    uint64(size * GIGABYTE),
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
					Shared:  true,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())
			Expect(volumeID).ToNot(BeEmpty())

			By("Checking if no of volumes present in cluster increases by 1")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)
		})

		// Has Issue with NFS driver
		// It allows to create a volume with blank name or double-spaces or n-spaces
		// Issue no #321
		if volumeDriver != "nfs" {

			It("Should fail to create volume with blank name", func() {

				By("Creating a volume blank name")
				var err error

				var size = 1
				vr := &api.VolumeCreateRequest{
					Locator: &api.VolumeLocator{
						Name: "",
						VolumeLabels: map[string]string{
							"class": "sanity-test-class",
						},
					},
					Source: &api.Source{},
					Spec: &api.VolumeSpec{
						Size:    uint64(size * GIGABYTE),
						HaLevel: 1,
					},
				}

				volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty volume name"))
			})
		}
	})

	Describe("Volume Inspect", func() {

		var (
			volumeID         string
			volumeIDs        []string
			numVolumesBefore int
			numVolumesAfter  int
			volumesToCreate  int
		)
		AfterEach(func() {
			var err error

			if volumesToCreate != 0 {
				for i := 0; i < len(volumeIDs); i++ {
					err = volumedriver.Delete(volumeIDs[i])
					Expect(err).ToNot(HaveOccurred())
				}

				volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
				Expect(err).ToNot(HaveOccurred())
				numVolumesAfter = len(volumes)
			}
		})

		It("Should create two volumes successfully", func() {

			By("Creating the volumes")
			var size = 2
			name := "inspect-vol-"

			volumesToCreate = 2
			for i := 0; i < volumesToCreate; i++ {

				By("Querying the number of volumes before create")

				volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
				numVolumesBefore = len(volumes)
				Expect(err).NotTo(HaveOccurred())

				vr := &api.VolumeCreateRequest{
					Locator: &api.VolumeLocator{
						Name: name + strconv.Itoa(i),
						VolumeLabels: map[string]string{
							"class": "sanity-test-class",
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
				Expect(volumeID).ToNot(BeNil())

				volumeIDs = append(volumeIDs, volumeID)

				By("Checking if volume created successfully with the provided params")
				testIfVolumeCreatedSuccessfully(volumedriver, volumeIDs[i], numVolumesBefore, vr)
			}
		})

		It("Should fail to inspect a volume", func() {

			// REST endpoint doesn't throw any error where cli throws an error
			By("Inspecting a volume that doesn't exist")
			volumesToCreate = 0
			volumes, err := volumedriver.Inspect([]string{"volume-id-doesnt-exist"})

			Expect(err).To(BeNil())
			Expect(volumes).To(BeEmpty())
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

			var size = 1
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-delete",
					VolumeLabels: map[string]string{
						"class": "cd-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    uint64(size * GIGABYTE),
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
			volumeIDs        []string
			volumesToCreate  = 3
		)

		AfterEach(func() {

			for _, id := range volumeIDs {
				err := volumedriver.Delete(id)
				Expect(err).ToNot(HaveOccurred())
			}

		})

		It("Should Enumerate all volumes ", func() {

			By("Creating a list of volumes")

			size := 3
			for i := 0; i < volumesToCreate; i++ {

				By("Querying the number of volumes before create")
				volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
				numVolumesBefore = len(volumes)
				Expect(err).NotTo(HaveOccurred())

				vr := &api.VolumeCreateRequest{
					Locator: &api.VolumeLocator{
						Name: "class-enumerate-" + strconv.Itoa(i),
						VolumeLabels: map[string]string{
							"class": "Class-A",
						},
					},
					Source: &api.Source{},
					Spec: &api.VolumeSpec{
						HaLevel: 1,
						Size:    uint64(size * GIGABYTE),
					},
				}

				volumeID, err := volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())
				Expect(err).NotTo(HaveOccurred())
				Expect(volumeID).ToNot(BeNil())

				volumeIDs = append(volumeIDs, volumeID)

				By("Checking if volume created successfully with the provided params")
				testIfVolumeCreatedSuccessfully(volumedriver, volumeIDs[i], numVolumesBefore, vr)
			}

			By("Enumerating Class-A volumes")

			volumes, err := volumedriver.Enumerate(
				&api.VolumeLocator{
					VolumeLabels: map[string]string{
						"class": "Class-A",
					},
				}, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(volumes)).To(BeEquivalentTo(volumesToCreate))
		})
	})

	Describe("Volume Attach Detach", func() {
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

		It("Should attach and detach successfully", func() {

			By("Creating the volume")

			var err error
			var size uint64 = 5
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-attach",
					VolumeLabels: map[string]string{
						"class": "attach-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    uint64(size * GIGABYTE),
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Attaching the volume to a node")

			str, err := volumedriver.Attach(volumeID, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(str).NotTo(BeNil())

			By("Inspecting the volume and checking attached_on field is not empty ")

			volumes, err := volumedriver.Inspect([]string{volumeID})

			Expect(err).NotTo(HaveOccurred())
			Expect(volumes[0].GetAttachedOn()).ToNot(BeEquivalentTo(""))
			Expect(volumes[0].GetAttachedState()).To(BeEquivalentTo(api.AttachState_ATTACH_STATE_EXTERNAL))

			By("Detaching the volume successfully")
			err = volumedriver.Detach(volumeID, nil)
			Expect(err).ToNot(HaveOccurred())

		})
	})

	Describe("Volume Mount Unmount", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			mountPath        string
			volumesToCreate  int
		)

		BeforeEach(func() {
			var err error
			mountPath = "/mnt"
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

		It("Should mount and unmount successfully", func() {

			By("Creating the volume")

			var err error

			volumesToCreate = 1
			var size = 5
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-mount",
					VolumeLabels: map[string]string{
						"class": "mount-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    uint64(size * GIGABYTE),
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Attaching a volume before mounting")

			str, err := volumedriver.Attach(volumeID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(str).NotTo(BeEmpty())

			By("Attaching mounting the volume to mount Path")

			err = volumedriver.Mount(volumeID, mountPath, nil)
			Expect(err).NotTo(HaveOccurred())

			By("Unmounting the volume successfully")
			err = volumedriver.Unmount(volumeID, mountPath, nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Volume Update", func() {

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

			By("Deleting the volume successfully")

			err = volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should update successfully with the new volume size.", func() {

			By("Creating the volume")

			var err error

			var size = 5
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-update-size",
					VolumeLabels: map[string]string{
						"class": "update-size-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    uint64(size * GIGABYTE),
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("First attaching and mounting the volume to a node")

			str, err := volumedriver.Attach(volumeID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(str).NotTo(BeNil())

			err = volumedriver.Mount(volumeID, "/mnt", nil)
			Expect(err).NotTo(HaveOccurred())

			By("Updating the volume spec with new random size.")
			newSize := random(size+1, 100)

			set := &api.VolumeSetRequest{
				Locator: vr.GetLocator(),
				Spec: &api.VolumeSpec{
					Size:             uint64(newSize * GIGABYTE),
					SnapshotInterval: math.MaxUint32,
				},
			}

			err = volumedriver.Set(volumeID, set.GetLocator(), set.GetSpec())
			Expect(err).NotTo(HaveOccurred())

			By("Inspecting the volume for new updates")

			volumes, err := volumedriver.Inspect([]string{volumeID})
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes[0].GetSpec().GetSize()).To(BeEquivalentTo(set.GetSpec().GetSize()))

			By("Detaching the volume and unmount successfully")

			err = volumedriver.Unmount(volumeID, "/mnt", nil)
			Expect(err).ToNot(HaveOccurred())

			err = volumedriver.Detach(volumeID, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should update volume successfully with new HA level", func() {

			By("Creating the volume")

			var err error
			var haLevel = 1

			var size = 5
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "vol-update-ha",
					VolumeLabels: map[string]string{
						"class": "update-ha-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:    uint64(size * GIGABYTE),
					HaLevel: int64(haLevel),
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			newHALevel := vr.Spec.HaLevel + 1
			By("Updating the volume HA level ")
			set := &api.VolumeSetRequest{
				Locator: vr.GetLocator(),
				Spec: &api.VolumeSpec{
					HaLevel: int64(newHALevel),
					ReplicaSet: &api.ReplicaSet{
						Nodes: []string{""},
					},
					SnapshotInterval: math.MaxUint32,
				},
			}

			err = volumedriver.Set(volumeID, set.Locator, set.Spec)
			Expect(err).NotTo(HaveOccurred())

			By("Inspecting the volume for new updates")
			time.Sleep(time.Second * 10)

			volumes, err := volumedriver.Inspect([]string{volumeID})
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes[0].Spec.HaLevel).To(BeEquivalentTo(newHALevel))
		})

	})

	Describe("Volume Stats [Stats]", func() {

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

		It("Should retrieve volume stats successfully", func() {

			By("Creating the volume")

			var err error
			var size = 5
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "get-stats",
					VolumeLabels: map[string]string{
						"class": "stat-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   2,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: 128,
					Cos:       api.CosType_LOW,
					Shared:    true,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Getting the stats")

			stats, err := volumedriver.Stats(volumeID, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(stats.String()).To(Not(BeNil()))
		})
	})

	Describe("Volume ActiveRequests", func() {

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

		It("Should get ActiveRequests successfully", func() {

			By("Creating the volume")

			var err error
			var size = 5
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "get-active-requests",
					VolumeLabels: map[string]string{
						"class": "active-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   2,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: 128,
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Getting the Active Requests")

			activeRequests, err := volumedriver.GetActiveRequests()
			Expect(err).NotTo(HaveOccurred())
			Expect(activeRequests).To(Not(BeNil()))
		})
	})

	Describe("Volume UsedSize", func() {

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

		It("Should get volume used size successfully", func() {

			By("Creating the volume")

			var err error
			var size = 2
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "get-used-size",
					VolumeLabels: map[string]string{
						"class": "usedSize-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   2,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: 128,
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())
			Expect(volumeID).ToNot(BeNil())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Getting the used size of the created volume")

			usedSize, err := volumedriver.UsedSize(volumeID)
			Expect(err).NotTo(HaveOccurred())
			Expect(usedSize).To(Not(BeNil()))
		})
	})

	Describe("Volume Quiesce Unquiesce", func() {

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
			Expect(err).NotTo(HaveOccurred())

			err = volumedriver.Detach(volumeID, nil)
			Expect(err).ToNot(HaveOccurred())

			err = volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should quiesce unquiesce volume successfully", func() {

			By("Creating the volume")

			var err error
			var size = 3
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "volume-quiesce",
					VolumeLabels: map[string]string{
						"class": "quiesce-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   2,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: 128,
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Queiscing the volume without Attaching")
			err = volumedriver.Quiesce(volumeID, 0, "")
			Expect(err).To(HaveOccurred())

			By("Now Attaching and mounting the volume")
			str, err := volumedriver.Attach(volumeID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(str).NotTo(BeNil())

			err = volumedriver.Mount(volumeID, "/mnt", nil)
			Expect(err).NotTo(HaveOccurred())

			By("Quiescing the volume")
			err = volumedriver.Quiesce(volumeID, 0, "")
			Expect(err).NotTo(HaveOccurred())

			By("Unquescing the quiesced volume")

			err = volumedriver.Unquiesce(volumeID)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
