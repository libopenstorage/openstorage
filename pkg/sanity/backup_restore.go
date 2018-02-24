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
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"

	"github.com/libopenstorage/openstorage/volume"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volume [Backup Restore Tests]", func() {
	var (
		restClient   *client.Client
		volumedriver volume.VolumeDriver
		credentials  *api.CredCreateRequest
		credUUID     string
	)

	BeforeEach(func() {

		if cloudBackupConfig == nil {
			Skip("Skipping cloud backup/restore tests")
		}
		var err error
		restClient, err = volumeclient.NewDriverClient(osdAddress, volumeDriver, volume.APIVersion, "")

		Expect(err).ToNot(HaveOccurred())
		volumedriver = volumeclient.VolumeDriver(restClient)

		By("Creating Credentials first")

		if cloudBackupConfig != nil {
			cloudParamsMap := cloudBackupConfig.CloudProviders["azure"]
			credentials = &api.CredCreateRequest{InputParams: cloudParamsMap}
		}

		credUUID, err = volumedriver.CredsCreate(credentials.InputParams)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {

		err := volumedriver.CredsDelete(credUUID)
		Expect(err).NotTo(HaveOccurred())

	})

	Describe("Volume Backup", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			bkpStatusReq     *api.BackupStsRequest
			bkpStatusResp    *api.BackupStsResponse
			bkpStatus        api.BackupStatus
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

		It("Should create Volume successfully for backup", func() {

			By("Creating the volume")

			var err error
			var size = 1
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-for-backup",
					VolumeLabels: map[string]string{
						"class": "backup-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   1,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: 128,
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Doing Backup")

			bkpReq := &api.BackupRequest{
				CredentialUUID: credUUID,
				Full:           false,
				VolumeID:       volumeID,
			}

			err = volumedriver.Backup(bkpReq)
			Expect(err).To((BeNil()))

			time.Sleep(time.Second * 10)
			By("Checking backup status")

			// timeout after 5 mins
			timeout := 300
			timespent := 0
			for timespent < timeout {
				bkpStatusReq = &api.BackupStsRequest{
					SrcVolumeID: volumeID,
				}
				bkpStatusResp = volumedriver.BackupStatus(bkpStatusReq)
				Expect(bkpStatusResp.StsErr).To(BeNil())

				bkpStatus = bkpStatusResp.Statuses[volumeID]
				if strings.Contains(bkpStatus.Status, "Done") {
					break
				}
				if strings.Contains(bkpStatus.Status, "Active") {
					time.Sleep(time.Second * 10)
					timeout += 10
				}
				if strings.Contains(bkpStatus.Status, "Failed") {
					break
				}
			}

			Expect(bkpStatus.Status).To(BeEquivalentTo("Done"))
		})
	})

	Describe("Volume Backup Enumerate", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			bkpStatusReq     *api.BackupStsRequest
			bkpStatusResp    *api.BackupStsResponse
			bkpStatus        api.BackupStatus
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

		It("Should create enumerate backup volumes", func() {

			By("Creating the volume")

			var err error
			var size = 1
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "enumerate-for-backup",
					VolumeLabels: map[string]string{
						"class": "enumerate-backp-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   1,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: 128,
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Doing Backup")

			bkpReq := &api.BackupRequest{
				CredentialUUID: credUUID,
				Full:           false,
				VolumeID:       volumeID,
			}

			err = volumedriver.Backup(bkpReq)
			Expect(err).To(BeNil())

			time.Sleep(time.Second * 10)

			By("Checking backup status")

			maxRetries := 3
			retries := 0

			for retries < maxRetries {
				bkpStatusReq = &api.BackupStsRequest{
					SrcVolumeID: volumeID,
				}
				bkpStatusResp = volumedriver.BackupStatus(bkpStatusReq)
				Expect(bkpStatusResp.StsErr).To(BeNil())

				bkpStatus = bkpStatusResp.Statuses[volumeID]
				if strings.Contains(bkpStatus.Status, "Done") {
					break
				}
				if strings.Contains(bkpStatus.Status, "Active") {
					time.Sleep(time.Second * 10)
				}
				if strings.Contains(bkpStatus.Status, "failed") {
					// give 3 attempts for backup to be successfull before declaring failed
					time.Sleep(time.Second * 10)
					retries++
				}
			}

			Expect(bkpStatus.Status).To(BeEquivalentTo("Done"))

			By("Backup enumerate")

			bkpEnumReq := &api.BackupEnumerateRequest{
				BackupGenericRequest: api.BackupGenericRequest{
					All:            false,
					CredentialUUID: credUUID,
					SrcVolumeID:    volumeID,
				},
			}

			enumResp := volumedriver.BackupEnumerate(bkpEnumReq)
			Expect(enumResp.EnumerateErr).To(BeNil())
			Expect(len(enumResp.Backups)).ToNot(BeNil())
			Expect(enumResp.Backups[0].SrcVolumeID).To(BeEquivalentTo(volumeID))
		})
	})

	Describe("Volume Backup Restore", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			bkpStatusReq     *api.BackupStsRequest
			bkpStatusResp    *api.BackupStsResponse
			bkpStatus        api.BackupStatus
			restoredVolume   string
		)

		BeforeEach(func() {
			var err error
			restoredVolume = "backup-restored"
			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
			numVolumesBefore = len(volumes)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			var err error

			err = volumedriver.Delete(volumeID)
			Expect(err).ToNot(HaveOccurred())

			err = volumedriver.Delete(restoredVolume)
			Expect(err).ToNot(HaveOccurred())

			volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, nil)
			Expect(err).ToNot(HaveOccurred())
			numVolumesAfter = len(volumes)
		})

		It("Should restore backup", func() {

			By("Creating the volume")

			var err error
			var size = 1
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "create-to-restore-backup",
					VolumeLabels: map[string]string{
						"class": "restore-class",
					},
				},
				Source: &api.Source{},
				Spec: &api.VolumeSpec{
					Size:      uint64(size * GIGABYTE),
					HaLevel:   1,
					Format:    api.FSType_FS_TYPE_EXT4,
					BlockSize: 128,
					Cos:       api.CosType_LOW,
				},
			}

			volumeID, err = volumedriver.Create(vr.GetLocator(), vr.GetSource(), vr.GetSpec())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			By("Doing Backup")

			bkpReq := &api.BackupRequest{
				CredentialUUID: credUUID,
				Full:           false,
				VolumeID:       volumeID,
			}

			time.Sleep(time.Second * 10)
			err = volumedriver.Backup(bkpReq)
			Expect(err).NotTo(HaveOccurred())

			By("Checking backup status")

			time.Sleep(time.Second * 10)
			maxRetries := 3
			retries := 0

			for retries < maxRetries {
				bkpStatusReq = &api.BackupStsRequest{
					SrcVolumeID: volumeID,
				}
				bkpStatusResp = volumedriver.BackupStatus(bkpStatusReq)

				//	Expect(bkpStatusResp.StsErr).To(BeNil())

				bkpStatus = bkpStatusResp.Statuses[volumeID]
				if strings.Contains(bkpStatus.Status, "Done") {
					break
				}
				if strings.Contains(bkpStatus.Status, "Active") {
					time.Sleep(time.Second * 10)
				}
				if strings.Contains(bkpStatus.Status, "Failed") {
					// give 3 attempts for backup to be successfull before declaring failed
					time.Sleep(time.Second * 10)
					retries++
				}
			}

			Expect(bkpStatus.Status).To(BeEquivalentTo("Done"))

			By("Backup enumerate")

			bkpEnumReq := &api.BackupEnumerateRequest{
				BackupGenericRequest: api.BackupGenericRequest{
					All:            false,
					CredentialUUID: credUUID,
					SrcVolumeID:    volumeID,
				},
			}

			enumResp := volumedriver.BackupEnumerate(bkpEnumReq)
			Expect(enumResp.EnumerateErr).To(BeEmpty())
			Expect(len(enumResp.Backups)).ToNot(BeNil())
			Expect(enumResp.Backups[0].SrcVolumeID).To(BeEquivalentTo(volumeID))

			By("Backup restore")

			// Get the backup cloud backupid from backupEnumerate
			bkpID := enumResp.Backups[0].BackupID

			bkpRestoreReq := &api.BackupRestoreRequest{
				CloudBackupID:     bkpID,
				CredentialUUID:    credUUID,
				RestoreVolumeName: restoredVolume,
			}
			bkpRestoreResp := volumedriver.BackupRestore(bkpRestoreReq)
			Expect(bkpRestoreResp.RestoreErr).To(BeEmpty())

			By("Inspecting the restored volume")

			volumes, err := volumedriver.Inspect([]string{restoredVolume})

			Expect(len(volumes)).To(BeEquivalentTo(1))
			Expect(volumes[0].Locator.Name).To(BeEquivalentTo(restoredVolume))
		})
	})
})
