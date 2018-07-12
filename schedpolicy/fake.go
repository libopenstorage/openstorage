/*
Package fake provides an in-memory fake implementation of the Scheduler
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
package schedpolicy

import (
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
	"path"
)

const (
	fakeSchedPrefix = "/fake/sched"
)

type fakeSchedMgr struct {
	kv kvdb.Kvdb
}

func NewFakeScheduler() *fakeSchedMgr {
	// This instance of the KVDB is Always in memory and created for each instance of the fake driver
	// It is not necessary to run a single instance, and it helps tests create a new kvdb on each test
	kv, err := kvdb.New(mem.Name, "fake_sched", []string{}, nil, logrus.Panicf)
	if err != nil {
		logrus.Fatalf("Failed to create kv: %v", err)
		return nil
	}
	return &fakeSchedMgr{
		kv: kv,
	}
}

func (s *fakeSchedMgr) SchedPolicyCreate(name, sched string) error {
	_, err := s.kv.Create(fakeSchedPrefix+"/"+name, sched, 0)
	return err
}

func (s *fakeSchedMgr) SchedPolicyUpdate(name, sched string) error {
	_, err := s.kv.Update(fakeSchedPrefix+"/"+name, sched, 0)
	return err
}

func (s *fakeSchedMgr) SchedPolicyDelete(name string) error {
	s.kv.Delete(fakeSchedPrefix + "/" + name)
	return nil
}

func (s *fakeSchedMgr) SchedPolicyEnumerate() ([]*SchedPolicy, error) {
	kvp, err := s.kv.Enumerate(fakeSchedPrefix)
	if err != nil {
		return nil, err
	}

	list := make([]*SchedPolicy, len(kvp))
	for i, kv := range kvp {
		list[i] = &SchedPolicy{
			Name:     path.Base(kv.Key),
			Schedule: string(kv.Value),
		}
	}

	return list, nil
}

func (s *fakeSchedMgr) SchedPolicyGet(name string) (*SchedPolicy, error) {
	kvp, err := s.kv.Get(fakeSchedPrefix + "/" + name)
	if err != nil {
		return nil, err
	}

	return &SchedPolicy{
		Name:     name,
		Schedule: string(kvp.Value),
	}, nil
}
