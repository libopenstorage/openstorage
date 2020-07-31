/*
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
package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		s                string
		expectedDuration time.Duration
		expectFail       bool
	}{
		{
			s:                "123y",
			expectedDuration: 123 * time.Hour * 24 * 365,
		},
		{
			s:                "123d",
			expectedDuration: 123 * time.Hour * 24,
		},
		{
			s:                "123h",
			expectedDuration: 123 * time.Hour,
		},
		{
			s:                "123m",
			expectedDuration: 123 * time.Minute,
		},
		{
			s:                "123s",
			expectedDuration: 123 * time.Second,
		},
		{
			s:                "123sabcd",
			expectedDuration: 0,
			expectFail:       true,
		},
		{
			s:                "ab123s",
			expectedDuration: 0,
			expectFail:       true,
		},
		{
			s:                "ab123s123",
			expectedDuration: 0,
			expectFail:       true,
		},
		{
			s:                "123",
			expectedDuration: 0,
			expectFail:       true,
		},
		{
			s:                "",
			expectedDuration: 0,
			expectFail:       true,
		},
		{
			s:                "12134212342342347239847asdasdf",
			expectedDuration: 0,
			expectFail:       true,
		},
	}

	for _, test := range tests {
		duration, err := ParseToDuration(test.s)
		if test.expectFail {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expectedDuration, duration, test.s)
		}
	}
}
