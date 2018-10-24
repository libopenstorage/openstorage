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
		credsUUIDMap map[string]string
	)

	BeforeEach(func() {

		credsUUIDMap = make(map[string]string)

		if cloudBackupConfig == nil {
			Skip("Skipping cloud backup/restore tests")
		}
		var err error
		restClient, err = volumeclient.NewDriverClient(osdAddress, volumeDriver, volume.APIVersion, "")

		Expect(err).ToNot(HaveOccurred())
		volumedriver = volumeclient.VolumeDriver(restClient)

		for provider, providerParams := range cloudBackupConfig.CloudProviders {

			By("Creating Credentials first")

			credentials = &api.CredCreateRequest{InputParams: providerParams}
			credUUID, err = volumedriver.CredsCreate(credentials.InputParams)
			Expect(err).NotTo(HaveOccurred())

			By("Validating credentials for the provider - " + provider)
			err = volumedriver.CredsValidate(credUUID)
			Expect(err).NotTo(HaveOccurred())

			if err == nil {
				credsUUIDMap[provider] = credUUID
			}
		}

		if len(credsUUIDMap) == 0 {
			Skip("Skipping cloud backup/restore tests as none of the credentials provided could be validated.")
		}
	})

	AfterEach(func() {

		for _, creduuid := range credsUUIDMap {
			err := volumedriver.CredsDelete(creduuid)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Volume Backup", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			bkpStatusReq     *api.CloudBackupStatusRequest
			bkpStatus        api.CloudBackupStatus
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
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			for provider, uuid := range credsUUIDMap {
				credUUID = uuid
				By("Doing Backup on " + provider)

				bkpReq := &api.CloudBackupCreateRequest{
					CredentialUUID: credUUID,
					Full:           false,
					VolumeID:       volumeID,
				}

				// Attaching the volume first
				str, err := volumedriver.Attach(volumeID, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				_, err = volumedriver.CloudBackupCreate(bkpReq)
				Expect(err).To((BeNil()))

				By("Checking backup status")

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.CloudBackupStatusRequest{
						SrcVolumeID: volumeID,
					}
					bkpStatusResp, err := volumedriver.CloudBackupStatus(bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volumeID]
					if bkpStatus.Status == api.CloudBackupStatusDone {
						break
					}
					if bkpStatus.Status == api.CloudBackupStatusActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.CloudBackupStatusFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.CloudBackupStatusDone))
			}
		})
	})

	Describe("Volume Backup Enumerate", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			bkpStatusReq     *api.CloudBackupStatusRequest
			bkpStatus        api.CloudBackupStatus
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
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			for provider, uuid := range credsUUIDMap {
				credUUID = uuid
				By("Doing Backup on " + provider)
				bkpReq := &api.CloudBackupCreateRequest{
					CredentialUUID: credUUID,
					Full:           false,
					VolumeID:       volumeID,
				}

				// Attaching the volume first
				str, err := volumedriver.Attach(volumeID, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				_, err = volumedriver.CloudBackupCreate(bkpReq)
				Expect(err).To(BeNil()) // give 3 attempts for backup to be successfull before declaring failed

				By("Checking backup status")

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.CloudBackupStatusRequest{
						SrcVolumeID: volumeID,
					}
					bkpStatusResp, err := volumedriver.CloudBackupStatus(bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volumeID]
					if bkpStatus.Status == api.CloudBackupStatusDone {
						break
					}
					if bkpStatus.Status == api.CloudBackupStatusActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.CloudBackupStatusFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.CloudBackupStatusDone))

				By("Backup enumerate")

				bkpEnumReq := &api.CloudBackupEnumerateRequest{
					CloudBackupGenericRequest: api.CloudBackupGenericRequest{
						All:            false,
						CredentialUUID: credUUID,
						SrcVolumeID:    volumeID,
					},
				}

				enumResp, err := volumedriver.CloudBackupEnumerate(bkpEnumReq)
				Expect(err).To(BeNil())
				Expect(len(enumResp.Backups)).ToNot(BeNil())
				Expect(enumResp.Backups[0].SrcVolumeID).To(BeEquivalentTo(volumeID))
			}
		})
	})

	Describe("Volume Backup Restore", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			bkpStatusReq     *api.CloudBackupStatusRequest
			bkpStatus        api.CloudBackupStatus
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

			err = volumedriver.Detach(volumeID, nil)
			Expect(err).ToNot(HaveOccurred())

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
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			for provider, uuid := range credsUUIDMap {
				credUUID = uuid
				By("Doing Backup on " + provider)

				bkpReq := &api.CloudBackupCreateRequest{
					CredentialUUID: credUUID,
					Full:           false,
					VolumeID:       volumeID,
				}

				// Attaching the volume first
				str, err := volumedriver.Attach(volumeID, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				_, err = volumedriver.CloudBackupCreate(bkpReq)
				Expect(err).NotTo(HaveOccurred())

				By("Checking backup status")

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.CloudBackupStatusRequest{
						SrcVolumeID: volumeID,
					}
					bkpStatusResp, err := volumedriver.CloudBackupStatus(bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volumeID]
					if bkpStatus.Status == api.CloudBackupStatusDone {
						break
					}
					if bkpStatus.Status == api.CloudBackupStatusActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.CloudBackupStatusFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.CloudBackupStatusDone))

				By("Backup enumerate")
				bkpEnumReq := &api.CloudBackupEnumerateRequest{
					CloudBackupGenericRequest: api.CloudBackupGenericRequest{
						All:            false,
						CredentialUUID: credUUID,
						SrcVolumeID:    volumeID,
					},
				}
				enumResp, err := volumedriver.CloudBackupEnumerate(bkpEnumReq)
				Expect(err).To(BeNil())
				Expect(len(enumResp.Backups)).ToNot(BeNil())
				Expect(enumResp.Backups[0].SrcVolumeID).To(BeEquivalentTo(volumeID))

				By("Backup restore")

				// Get the backup cloud backupid from backupEnumerate
				bkpID := enumResp.Backups[0].ID

				bkpRestoreReq := &api.CloudBackupRestoreRequest{
					ID:                bkpID,
					CredentialUUID:    credUUID,
					RestoreVolumeName: restoredVolume,
				}
				bkpRestoreResp, err := volumedriver.CloudBackupRestore(bkpRestoreReq)
				Expect(err).To(BeNil())

				By("Inspecting the restored volume")

				volumes, err := volumedriver.Inspect([]string{bkpRestoreResp.RestoreVolumeID})
				Expect(err).NotTo(HaveOccurred())
				Expect(len(volumes)).To(BeEquivalentTo(1))
				Expect(volumes[0].Locator.Name).To(BeEquivalentTo(restoredVolume))
			}
		})
	})

	Describe("Volume Backup Schedule Create", func() {

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

		It("Should create a backup schedule ", func() {

			By("Creating the volume")

			var err error
			var size = 1
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "schedule-create",
					VolumeLabels: map[string]string{
						"class": "schedule-class",
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
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			for _, uuid := range credsUUIDMap {
				credUUID = uuid
				By("Creating a backup schedule")

				bkpScheduleInfo := &api.CloudBackupSchedCreateRequest{
					CloudBackupScheduleInfo: api.CloudBackupScheduleInfo{Schedule: "- freq: daily\n  hour: 23\n  minute: 00",
						CredentialUUID: credUUID,
						MaxBackups:     1,
						SrcVolumeID:    volumeID,
					},
				}
				bkpScheduleResponse, err := volumedriver.CloudBackupSchedCreate(bkpScheduleInfo)
				Expect(err).To(BeNil())
				Expect(bkpScheduleResponse.UUID).NotTo(BeNil())
			}
		})
	})

	Describe("Volume Backup Schedule Delete", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			schedules        []string
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

		It("Should delete a backup schedule ", func() {

			By("Creating the volume")

			var err error
			var size = 1
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "schedule-delete",
					VolumeLabels: map[string]string{
						"class": "schedule-class",
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
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			for _, uuid := range credsUUIDMap {
				credUUID = uuid
				By("Creating a backup schedule")

				bkpScheduleInfo := &api.CloudBackupSchedCreateRequest{
					CloudBackupScheduleInfo: api.CloudBackupScheduleInfo{
						Schedule:       "- freq: daily\n  hour: 23\n  minute: 00",
						CredentialUUID: credUUID,
						MaxBackups:     1,
						SrcVolumeID:    volumeID,
					},
				}
				bkpScheduleResponse, err := volumedriver.CloudBackupSchedCreate(bkpScheduleInfo)
				Expect(err).To(BeNil())
				Expect(bkpScheduleResponse.UUID).NotTo(BeEmpty())

				schedules = append(schedules, bkpScheduleResponse.UUID)
			}

			By("Deleting the created schedules")

			for _, schedUUID := range schedules {

				bkpScheduleDeleteReq := &api.CloudBackupSchedDeleteRequest{
					UUID: schedUUID,
				}
				err = volumedriver.CloudBackupSchedDelete(bkpScheduleDeleteReq)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Describe("Volume Backup Schedule Enumerate", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			schedules        []string
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

		It("Should enumerate a backup schedule ", func() {

			By("Creating the volume")

			var err error
			var size = 1
			vr := &api.VolumeCreateRequest{
				Locator: &api.VolumeLocator{
					Name: "schedule-enumerate",
					VolumeLabels: map[string]string{
						"class": "schedule-class",
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
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			for _, uuid := range credsUUIDMap {
				credUUID = uuid
				By("Creating a backup schedule")

				bkpScheduleInfo := &api.CloudBackupSchedCreateRequest{
					CloudBackupScheduleInfo: api.CloudBackupScheduleInfo{
						Schedule:       "- freq: daily\n  hour: 23\n  minute: 00",
						CredentialUUID: credUUID,
						MaxBackups:     1,
						SrcVolumeID:    volumeID,
					},
				}
				bkpScheduleResponse, err := volumedriver.CloudBackupSchedCreate(bkpScheduleInfo)
				Expect(err).To(BeNil())
				Expect(bkpScheduleResponse.UUID).NotTo(BeNil())

				schedules = append(schedules, bkpScheduleResponse.UUID)
			}

			By("Enumerating the created schedules")

			bkpScheduleEnumerateResp, err := volumedriver.CloudBackupSchedEnumerate()
			Expect(err).To(BeNil())
			Expect(len(bkpScheduleEnumerateResp.Schedules)).NotTo(BeZero())
		})
	})

	Describe("Volume Backup History & Catalogue", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volumeID         string
			bkpStatusReq     *api.CloudBackupStatusRequest
			bkpStatusResp    *api.CloudBackupStatusResponse
			bkpStatus        api.CloudBackupStatus
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
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")
			testIfVolumeCreatedSuccessfully(volumedriver, volumeID, numVolumesBefore, vr)

			for provider, uuid := range credsUUIDMap {
				credUUID = uuid
				By("Doing Backup on " + provider)

				bkpReq := &api.CloudBackupCreateRequest{
					CredentialUUID: credUUID,
					Full:           false,
					VolumeID:       volumeID,
				}

				// Attaching the volume first
				str, err := volumedriver.Attach(volumeID, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				_, err = volumedriver.CloudBackupCreate(bkpReq)
				Expect(err).To((BeNil()))

				By("Checking backup status")

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.CloudBackupStatusRequest{
						SrcVolumeID: volumeID,
					}
					bkpStatusResp, err = volumedriver.CloudBackupStatus(bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volumeID]
					if bkpStatus.Status == api.CloudBackupStatusDone {
						break
					}
					if bkpStatus.Status == api.CloudBackupStatusActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.CloudBackupStatusFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.CloudBackupStatusDone))

				By("Backup enumerate")
				bkpEnumReq := &api.CloudBackupEnumerateRequest{
					CloudBackupGenericRequest: api.CloudBackupGenericRequest{
						All:            false,
						CredentialUUID: credUUID,
						SrcVolumeID:    volumeID,
					},
				}
				enumResp, err := volumedriver.CloudBackupEnumerate(bkpEnumReq)
				Expect(err).To(BeNil())
				Expect(len(enumResp.Backups)).ToNot(BeNil())
				Expect(enumResp.Backups[0].SrcVolumeID).To(BeEquivalentTo(volumeID))

				// Get the backup cloud backupid from backupEnumerate
				bkpID := enumResp.Backups[0].ID

				By("Getting backup catalogue")

				bkpCatalogReq := &api.CloudBackupCatalogRequest{
					ID:             bkpID,
					CredentialUUID: credUUID,
				}

				bkpCatalogResp, err := volumedriver.CloudBackupCatalog(bkpCatalogReq)
				Expect(err).To(BeNil())
				Expect(bkpCatalogResp.Contents).To(Not(BeEmpty()))
			}

			By("Getting the Backup history")

			bkpHistory := &api.CloudBackupHistoryRequest{
				SrcVolumeID: volumeID,
			}

			bkpHistoryResp, err := volumedriver.CloudBackupHistory(bkpHistory)
			Expect(err).To(BeNil())

			// Expect the created backup to have an entry in the backup history
			isPresent := false
			for _, historyItem := range bkpHistoryResp.HistoryList {
				if historyItem.SrcVolumeID == volumeID {
					Expect(historyItem.Status).To(ContainSubstring("Cloudsnap Backup completed successfully"))
					isPresent = true
				}
			}
			Expect(isPresent).To(BeTrue())
		})
	})
})
