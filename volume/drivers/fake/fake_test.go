/*
Package fake provides an in-memory fake driver implementation
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
package fake

import (
	"testing"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	kv, err := kvdb.New(mem.Name, "fake_test", []string{}, nil, logrus.Panicf)
	if err != nil {
		logrus.Panicf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		logrus.Panicf("Failed to set KVDB instance")
	}
}

func TestFakeName(t *testing.T) {
	d, err := Init(map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, Name, d.Name())
}

func TestFakeCredentials(t *testing.T) {
	d, err := Init(map[string]string{})
	assert.NoError(t, err)

	id, err := d.CredsCreate(map[string]string{
		"hello": "world",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	creds, err := d.CredsEnumerate()
	assert.NoError(t, err)
	assert.NotEmpty(t, creds)
	assert.Len(t, creds, 1)

	data := creds[id]
	value, ok := data.(map[string]string)
	assert.True(t, ok)
	assert.NotEmpty(t, value)
	assert.Equal(t, value["hello"], "world")

	err = d.CredsDelete(id)
	assert.NoError(t, err)

	creds, err = d.CredsEnumerate()
	assert.NoError(t, err)
	assert.Empty(t, creds)
}
