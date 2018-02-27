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
	"math/rand"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	BYTE = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

func testIfVolumeCreatedSuccessfully(
	volumedriver volume.VolumeDriver,
	volumeID string,
	numVolumesBefore int,
	vr *api.VolumeCreateRequest) {

	Expect(volumeID).ToNot(BeNil())

	volumes, err := volumedriver.Enumerate(&api.VolumeLocator{}, make(map[string]string))
	Expect(err).NotTo(HaveOccurred())

	numVolumesAfter := len(volumes)

	Expect(numVolumesAfter).To(Equal(numVolumesBefore + 1))

	By("Inspecting the created volume")

	inspectVolumes := []string{volumeID}
	volumesList, err := volumedriver.Inspect(inspectVolumes)
	Expect(err).NotTo(HaveOccurred())
	Expect(volumesList).NotTo(BeEmpty())
	Expect(len(volumesList)).Should(BeEquivalentTo(1))
	Expect(volumesList[0].GetId()).Should(BeEquivalentTo(volumeID))

	// check volume specs
	Expect(volumesList[0].GetSpec().GetEphemeral()).To(BeEquivalentTo(vr.GetSpec().GetEphemeral()))
	//Expect(volumesList[0].GetSpec().GetAggregationLevel()).To(BeEquivalentTo(vr.GetSpec().GetAggregationLevel()))
	//Expect(volumesList[0].GetSpec().GetBlockSize()).To(BeEquivalentTo(vr.GetSpec().GetBlockSize()))
	Expect(volumesList[0].GetSpec().GetCascaded()).To(BeEquivalentTo(vr.GetSpec().GetCascaded()))
	Expect(volumesList[0].GetSpec().GetCompressed()).To(BeEquivalentTo(vr.GetSpec().GetCompressed()))
	//Expect(volumesList[0].GetSpec().GetCos()).To(BeEquivalentTo(vr.GetSpec().GetCos()))
	Expect(volumesList[0].GetSpec().GetDedupe()).To(BeEquivalentTo(vr.GetSpec().GetDedupe()))
	Expect(volumesList[0].GetSpec().GetEncrypted()).To(BeEquivalentTo(vr.GetSpec().GetEncrypted()))
	Expect(volumesList[0].GetSpec().GetFormat()).To(BeEquivalentTo(vr.GetSpec().GetFormat()))
	Expect(volumesList[0].GetSpec().GetGroup()).To(BeEquivalentTo(vr.GetSpec().GetGroup()))
	Expect(volumesList[0].GetSpec().GetGroupEnforced()).To(BeEquivalentTo(vr.GetSpec().GetGroupEnforced()))
	Expect(volumesList[0].GetSpec().GetHaLevel()).To(BeEquivalentTo(vr.GetSpec().GetHaLevel()))
	Expect(volumesList[0].GetSpec().GetIoProfile()).To(BeEquivalentTo(vr.GetSpec().GetIoProfile()))
	Expect(volumesList[0].GetSpec().GetJournal()).To(BeEquivalentTo(vr.GetSpec().GetJournal()))
	Expect(volumesList[0].GetSpec().GetNfs()).To(BeEquivalentTo(vr.GetSpec().GetNfs()))
	Expect(volumesList[0].GetSpec().GetPassphrase()).To(BeEquivalentTo(vr.GetSpec().GetPassphrase()))
	Expect(volumesList[0].GetSpec().GetReplicaSet()).To(BeEquivalentTo(vr.GetSpec().GetReplicaSet()))
	Expect(volumesList[0].GetSpec().GetScale()).To(BeEquivalentTo(vr.GetSpec().GetScale()))
	Expect(volumesList[0].GetSpec().GetShared()).To(BeEquivalentTo(vr.GetSpec().GetShared()))
	Expect(volumesList[0].GetSpec().GetSize()).To(BeEquivalentTo(vr.GetSpec().GetSize()))
	Expect(volumesList[0].GetSpec().GetSnapshotInterval()).To(BeEquivalentTo(vr.GetSpec().GetSnapshotInterval()))
	Expect(volumesList[0].GetSpec().GetSnapshotSchedule()).To(BeEquivalentTo(vr.GetSpec().GetSnapshotSchedule()))
	Expect(volumesList[0].GetSpec().GetSticky()).To(BeEquivalentTo(vr.GetSpec().GetSticky()))
}

//Returns an in between min and max. Min - included, Max excluded. So mathematically [min, max)
func random(min, max int) int {
	if max == min {
		return max
	}
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
