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
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"

	"github.com/libopenstorage/openstorage/volume"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credentials Tests", func() {
	var (
		restClient   *client.Client
		volumedriver volume.VolumeDriver
		credentials  *api.CredCreateRequest
	)

	BeforeEach(func() {

		if cloudBackupConfig == nil {
			Skip("Skipping credentials tests")
		}
		var err error
		restClient, err = volumeclient.NewDriverClient(osdAddress, volumeDriver, volume.APIVersion, "")

		Expect(err).ToNot(HaveOccurred())
		volumedriver = volumeclient.VolumeDriver(restClient)
	})

	Describe("Credentials Create", func() {

		var (
			credUUID     string
			credsUUIDMap map[string]string
		)

		AfterEach(func() {
			for _, creduuid := range credsUUIDMap {
				err := volumedriver.CredsDelete(creduuid)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		var err error

		It("Should Create credentials successfully", func() {
			credsUUIDMap = make(map[string]string)

			for provider, providerParams := range cloudBackupConfig.CloudProviders {
				By("Creating Credentials for - " + provider)

				credentials = &api.CredCreateRequest{InputParams: providerParams}
				credUUID, err = volumedriver.CredsCreate(credentials.InputParams)

				Expect(err).NotTo(HaveOccurred())
				Expect(credUUID).NotTo(BeEmpty())

				credsUUIDMap[provider] = credUUID
			}
		})
	})

	Describe("Credentials Enumerate and Delete", func() {

		var (
			numOfCredentialsToCreate = 3
			credsBefore              int
			credsAfter               int
			credsEnumerateMap        map[string]interface{}
			creds                    []string
		)

		AfterEach(func() {

			for _, creduuid := range creds {
				err := volumedriver.CredsDelete(creduuid)
				Expect(err).NotTo(HaveOccurred())
			}

			By("Enumerating the credentials after Delete")
			credsMap, err := volumedriver.CredsEnumerate()
			Expect(err).NotTo(HaveOccurred())
			credsAfter = len(credsMap)

			Expect(credsAfter).To(BeEquivalentTo(credsBefore))
		})

		It("Should enumerate created credentials successfully", func() {

			var err error

			By("First Enumerating the credentials before create")
			credsEnumerateMap, err = volumedriver.CredsEnumerate()
			Expect(err).NotTo(HaveOccurred())
			credsBefore = len(credsEnumerateMap)

			By("Creating multiple Credentials from the config file")

			for i := 0; i < numOfCredentialsToCreate; i++ {

				for provider, providerParams := range cloudBackupConfig.CloudProviders {
					By("Creating Credentials for - " + provider)

					credentials = &api.CredCreateRequest{InputParams: providerParams}
					credUUID, err := volumedriver.CredsCreate(credentials.InputParams)

					Expect(err).NotTo(HaveOccurred())
					Expect(credUUID).NotTo(BeEmpty())

					creds = append(creds, credUUID)
				}
			}

			By("Enumerating the credentials after Create")
			credsEnumerateMap, err = volumedriver.CredsEnumerate()
			Expect(err).NotTo(HaveOccurred())
			credsAfter = len(credsEnumerateMap)

			numOfProviders := len(cloudBackupConfig.CloudProviders)
			Expect(credsAfter).To(BeEquivalentTo(credsBefore + numOfCredentialsToCreate*numOfProviders))
		})
	})

	Describe("Credentials Validate", func() {
		var (
			credUUID     string
			credsUUIDMap map[string]string
		)

		AfterEach(func() {
			for _, creduuid := range credsUUIDMap {
				err := volumedriver.CredsDelete(creduuid)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should Validate the created credentials", func() {
			var err error
			credsUUIDMap = make(map[string]string)

			for provider, providerParams := range cloudBackupConfig.CloudProviders {
				By("Creating Credentials for - " + provider)

				credentials = &api.CredCreateRequest{InputParams: providerParams}
				credUUID, err = volumedriver.CredsCreate(credentials.InputParams)
				Expect(err).NotTo(HaveOccurred())
				Expect(credUUID).NotTo(BeEmpty())

				credsUUIDMap[provider] = credUUID
			}

			By("Validating the created credentials")
			for _, creduuid := range credsUUIDMap {
				err := volumedriver.CredsValidate(creduuid)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
