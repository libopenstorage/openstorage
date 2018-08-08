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

var _ = Describe("Volume [Snapshot Tests]", func() {
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

	Describe("Volume Snapshot Create", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			snapID           string
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

			err = volumedriver.Delete(snapID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should create Volume successfully for snapshot", func() {

			By("Creating the volume")

			var err error
			var size = 3
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-for-snapshot",
					VolumeLabels: map[string]string{
						"class": "create-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   1,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: int64(512 * KILOBYTE),
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Creating a snapshot based on the created volume")

			loc := &api.VolumeLocator{
				Name: "snapshot-of-" + volumeID,
			}

			snapID, err = volumedriver.Snapshot(volumeID, true, loc, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapID).To(Not(BeNil()))

			By("Checking the Parent field of the created snapshot")

			volumes, err := volumedriver.Inspect([]string{loc.GetName()})
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes).NotTo(BeEmpty())

			Expect(volumes[0].GetSource().GetParent()).To(BeEquivalentTo(volumeID))
			Expect(volumes[0].GetReadonly()).To(BeTrue())
		})
	})

	Describe("Volume Snapshot Enumerate", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			snapID           string
			snapIDs          []string
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

			for _, snapID := range snapIDs {
				err = volumedriver.Delete(snapID)
				Expect(err).ToNot(HaveOccurred())
			}

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should enumerate Volume snapshots", func() {

			By("Creating the volume")

			var err error
			var size = 3
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-for-snapshot",
					VolumeLabels: map[string]string{
						"class": "create-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   1,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: int64(32 * KILOBYTE),
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Creating a multiple [3] snapshots based on the created volume")

			numOfSnaps := 3

			for i := 0; i < numOfSnaps; i++ {

				loc := &api.VolumeLocator{
					Name: "snapshot-" + strconv.Itoa(i) + "-of-" + volumeID,
				}

				snapID, err = volumedriver.Snapshot(volumeID, true, loc, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(snapID).To(Not(BeNil()))

				snapIDs = append(snapIDs, snapID)
			}

			By("Checking the Parent field of the created snapshot")

			volumes, err := volumedriver.Inspect(snapIDs)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes).NotTo(BeEmpty())

			for _, snap := range volumes {

				Expect(snap.GetSource().GetParent()).To(BeEquivalentTo(volumeID))
				Expect(snap.GetReadonly()).To(BeTrue())
			}

			By("Enumerating the snapshots with the volumeID")

			allSnapsOfVolumeID, err := volumedriver.SnapEnumerate([]string{volumeID}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(allSnapsOfVolumeID)).To(BeEquivalentTo(numOfSnaps))
		})
	})

	Describe("Volume Snapshot Restore", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			snapID           string
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

			err = volumedriver.Delete(snapID)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should restore Volume successfully for snapshot", func() {

			By("Creating the volume")

			var err error
			var size = 3
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-restore",
					VolumeLabels: map[string]string{
						"class": "restore-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   1,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: int64(512 * KILOBYTE),
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Creating a snapshot based on the created volume")

			loc := &api.VolumeLocator{
				Name: "snapshot-of-" + volumeID,
			}

			snapID, err = volumedriver.Snapshot(volumeID, true, loc, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapID).To(Not(BeNil()))

			By("Checking the Parent field of the created snapshot")

			volumes, err := volumedriver.Inspect([]string{loc.GetName()})
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes).NotTo(BeEmpty())

			Expect(volumes[0].GetSource().GetParent()).To(BeEquivalentTo(volumeID))
			Expect(volumes[0].GetReadonly()).To(BeTrue())

			By("Restoring the volume from snapshot")

			err = volumedriver.Restore(volumeID, snapID)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
