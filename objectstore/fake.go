/*
Provides an in-memory fake implementation of the Objectstore
Copyright 2018 Portworx

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
package objectstore

import (
	"strings"

	"github.com/libopenstorage/openstorage/api"
	"github.com/pborman/uuid"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
)

const (
	fakeObjectstorePrefix = "/fake/objectstore"
)

type fakeObjectstoreMgr struct {
	kv kvdb.Kvdb
}

func NewfakeObjectstore() *fakeObjectstoreMgr {
	// This instance of the KVDB is Always in memory and created for each instance of the fake driver
	// It is not necessary to run a single instance, and it helps tests create a new kvdb on each test
	kv, err := kvdb.New(mem.Name, "fake_objectstore", []string{}, nil, logrus.Panicf)
	if err != nil {
		logrus.Fatalf("Failed to create kv: %v", err)
		return nil
	}

	return &fakeObjectstoreMgr{
		kv: kv,
	}
}

func (s *fakeObjectstoreMgr) ObjectStoreCreate(volumeID string) (*api.ObjectstoreInfo, error) {
	objectstoreID := strings.TrimSuffix(uuid.New(), "\n")
	fakeObjStoreInfo := &api.ObjectstoreInfo{
		Uuid:            objectstoreID,
		VolumeId:        volumeID,
		Enabled:         false,
		CurrentEndpoint: "http://test:9000",
	}

	_, err := s.kv.Create(fakeObjectstorePrefix+"/"+objectstoreID, fakeObjStoreInfo, 0)
	if err != nil {
		return nil, err
	}

	return fakeObjStoreInfo, nil
}

func (s *fakeObjectstoreMgr) ObjectStoreDelete(objectstoreID string) error {
	s.kv.Delete(fakeObjectstorePrefix + "/" + objectstoreID)
	return nil
}

func (s *fakeObjectstoreMgr) ObjectStoreUpdate(objectstoreID string, enable bool) error {
	var objInfo *api.ObjectstoreInfo
	_, err := s.kv.GetVal(fakeObjectstorePrefix+"/"+objectstoreID, &objInfo)
	if err != nil {
		return err
	}

	objInfo.Enabled = enable
	_, err = s.kv.Update(fakeObjectstorePrefix+"/"+objectstoreID, objInfo, 0)

	return err
}

func (s *fakeObjectstoreMgr) ObjectStoreInspect(objectstoreID string) (*api.ObjectstoreInfo, error) {
	var objInfo *api.ObjectstoreInfo
	_, err := s.kv.GetVal(fakeObjectstorePrefix+"/"+objectstoreID, &objInfo)
	if err != nil {
		return nil, err
	}

	return objInfo, nil
}
