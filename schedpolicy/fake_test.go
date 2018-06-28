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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFakeSchedule(t *testing.T) {
	f := NewFakeScheduler()
	err := f.SchedPolicyCreate("hello", "world")
	assert.NoError(t, err)
	err = f.SchedPolicyCreate("name", "sched")
	assert.NoError(t, err)

	// Update
	err = f.SchedPolicyUpdate("hello", "universe")
	assert.NoError(t, err)

	// Enumerate
	list, err := f.SchedPolicyEnumerate()
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Contains(t, list, &SchedPolicy{
		Name:     "hello",
		Schedule: "universe",
	})
	assert.Contains(t, list, &SchedPolicy{
		Name:     "name",
		Schedule: "sched",
	})

	// Delete
	err = f.SchedPolicyDelete("hello")
	assert.NoError(t, err)

	list, err = f.SchedPolicyEnumerate()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Contains(t, list, &SchedPolicy{
		Name:     "name",
		Schedule: "sched",
	})

	// Inspect
	policy, err := f.SchedPolicyGet("name")
	assert.NoError(t, err)
	assert.NotNil(t, policy)
	assert.Equal(t, policy.Name, "name")
	assert.Equal(t, policy.Schedule, "sched")
}
