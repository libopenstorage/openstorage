/*
Package api contains the external OpenStorage apis
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
package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudBackupStatusTypeToSdkCloudBackupStatusType(t *testing.T) {

	tests := []struct {
		internalType CloudBackupStatusType
		sdkType      SdkCloudBackupStatusType
	}{
		{
			CloudBackupStatusNotStarted,
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeNotStarted,
		},
		{
			CloudBackupStatusDone,
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone,
		},
		{
			CloudBackupStatusAborted,
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeAborted,
		},
		{
			CloudBackupStatusPaused,
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypePaused,
		},
		{
			CloudBackupStatusStopped,
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeStopped,
		},
		{
			CloudBackupStatusActive,
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive,
		},
		{
			CloudBackupStatusFailed,
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed,
		},
		{
			"What?",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeUnknown,
		},
	}

	for _, test := range tests {
		assert.Equal(t,
			test.sdkType,
			CloudBackupStatusTypeToSdkCloudBackupStatusType(test.internalType))
	}
}

func TestStringToSdkCloudBackupStatusType(t *testing.T) {

	tests := []struct {
		internalType string
		sdkType      SdkCloudBackupStatusType
	}{
		{
			"NotStarted",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeNotStarted,
		},
		{
			"Done",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone,
		},
		{
			"Aborted",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeAborted,
		},
		{
			"Paused",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypePaused,
		},
		{
			"Stopped",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeStopped,
		},
		{
			"Active",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive,
		},
		{
			"Failed",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed,
		},
		{
			"What?",
			SdkCloudBackupStatusType_SdkCloudBackupStatusTypeUnknown,
		},
	}

	for _, test := range tests {
		assert.Equal(t,
			test.sdkType,
			StringToSdkCloudBackupStatusType(test.internalType))
	}
}
