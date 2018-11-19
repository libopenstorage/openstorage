/*
Package sdk is the gRPC implementation of the SDK gRPC server
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

package sdk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/proto/time"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// configs is a test configuration list that we iterate over and test each case.
	configs = []struct {
		// name of the test, helps debugging when a certain test case fails.
		name string
		// req is the request object being sent over gRPC call.
		req *api.SdkAlertsEnumerateRequest
		// expected is the number of alerts objects returned in output.
		expected int
	}{
		{
			name: "ResourceTypeQuery, opt count",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewCountSpanOption(0, 10),
							},
						},
					},
				},
			},
			expected: 4,
		},
		{
			name: "ResourceTypeQuery, minsev none",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewMinSeverityOption(api.SeverityType_SEVERITY_TYPE_NONE),
							},
						},
					},
				},
			},
			expected: 4,
		},
		{
			name: "ResourceTypeQuery, minsev notify",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewMinSeverityOption(api.SeverityType_SEVERITY_TYPE_NOTIFY),
							},
						},
					},
				},
			},
			expected: 4,
		},
		{
			name: "ResourceTypeQuery, minsev warning",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewMinSeverityOption(api.SeverityType_SEVERITY_TYPE_WARNING),
							},
						},
					},
				},
			},
			expected: 3,
		},
		{
			name: "ResourceTypeQuery, minsev alarm",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewMinSeverityOption(api.SeverityType_SEVERITY_TYPE_ALARM),
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "ResourceTypeQuery, time range 1",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewTimeSpanOption(
									time.Now().AddDate(0, -3, 0),
									time.Now()),
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "ResourceTypeQuery, time range 2",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewTimeSpanOption(
									time.Now().AddDate(0, -6, 0),
									time.Now()),
							},
						},
					},
				},
			},
			expected: 4,
		},
		{
			name: "ResourceTypeQuery, time range 3",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewTimeSpanOption(
									time.Now().AddDate(0, -6, 0),
									time.Now().AddDate(0, -3, 0),
								),
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "ResourceTypeQuery, time range and severity warning",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewTimeSpanOption(
									time.Now().AddDate(0, -6, 0),
									time.Now().AddDate(0, -3, 0),
								),
							},
							{
								Opt: testNewMinSeverityOption(api.SeverityType_SEVERITY_TYPE_WARNING),
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "AlertTypeQuery, time range and severity warning",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewAlertTypeQuery(10, api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewTimeSpanOption(
									time.Now().AddDate(0, -6, 0),
									time.Now(),
								),
							},
							{
								Opt: testNewMinSeverityOption(api.SeverityType_SEVERITY_TYPE_WARNING),
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "ResourceIdQuery, time range and severity warning",
			req: &api.SdkAlertsEnumerateRequest{
				Queries: []*api.SdkAlertsQuery{
					{
						Query: testNewResourceIdQuery("maya", 10, api.ResourceType_RESOURCE_TYPE_DRIVE),
						Opts: []*api.SdkAlertsOption{
							{
								Opt: testNewTimeSpanOption(
									time.Now().AddDate(0, -6, 0),
									time.Now(),
								),
							},
							{
								Opt: testNewMinSeverityOption(api.SeverityType_SEVERITY_TYPE_WARNING),
							},
						},
					},
				},
			},
			expected: 1,
		},
	}
)

// testNewResourceTypeQuery provides a way to create alerts query object required
// for using alerts functionality in SDK.
func testNewResourceTypeQuery(resourceType api.ResourceType) *api.SdkAlertsQuery_ResourceTypeQuery {
	return &api.SdkAlertsQuery_ResourceTypeQuery{
		ResourceTypeQuery: &api.SdkAlertsResourceTypeQuery{
			ResourceType: resourceType,
		},
	}
}

// testNewAlertTypeQuery provides a way to create alerts query object required
// for using alerts functionality in SDK.
func testNewAlertTypeQuery(alertType int64, resourceType api.ResourceType) *api.SdkAlertsQuery_AlertTypeQuery {
	return &api.SdkAlertsQuery_AlertTypeQuery{
		AlertTypeQuery: &api.SdkAlertsAlertTypeQuery{
			ResourceType: resourceType,
			AlertType:    alertType,
		},
	}
}

// testNewResourceIdQuery provides a way to create alerts query object required
// for using alerts functionality in SDK.
func testNewResourceIdQuery(resourceId string, alertType int64, resourceType api.ResourceType) *api.SdkAlertsQuery_ResourceIdQuery {
	return &api.SdkAlertsQuery_ResourceIdQuery{
		ResourceIdQuery: &api.SdkAlertsResourceIdQuery{
			ResourceType: resourceType,
			AlertType:    alertType,
			ResourceId:   resourceId,
		},
	}
}

// testNewTimeSpanOption provides a way to create alerts options
func testNewTimeSpanOption(startTime, endTime time.Time) *api.SdkAlertsOption_TimeSpan {
	return &api.SdkAlertsOption_TimeSpan{
		TimeSpan: &api.SdkAlertsTimeSpan{
			StartTime: prototime.TimeToTimestamp(
				startTime),
			EndTime: prototime.TimeToTimestamp(
				endTime),
		},
	}
}

// testNewCountSpanOption provides a way to create alerts options
func testNewCountSpanOption(minCount, maxCount int64) *api.SdkAlertsOption_CountSpan {
	return &api.SdkAlertsOption_CountSpan{
		CountSpan: &api.SdkAlertsCountSpan{
			MinCount: minCount,
			MaxCount: maxCount,
		},
	}
}

// testNewMinSeverityOption provides a way to create alerts options
func testNewMinSeverityOption(minSev api.SeverityType) *api.SdkAlertsOption_MinSeverityType {
	return &api.SdkAlertsOption_MinSeverityType{
		MinSeverityType: minSev}
}

// testNewIsClearedOption provides a way to create alerts options
func testNewIsClearedOption(isCleared bool) *api.SdkAlertsOption_IsCleared {
	return &api.SdkAlertsOption_IsCleared{
		IsCleared: isCleared}
}

// TestAlertsServerEnumerate tests enumerate functionality over gRPC using mock.
func TestAlertsServerEnumerate(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageAlertsClient(s.Conn())

	for _, config := range configs {
		var filters []interface{}
		for _, filter := range getFilters(config.req.Queries) {
			filters = append(filters, filter)
		}
		myAlerts := make([]*api.Alert, config.expected)
		for i := range myAlerts {
			myAlerts[i] = new(api.Alert)
		}
		s.MockFilterDeleter().EXPECT().Enumerate(filters...).Return(myAlerts, nil).Times(1)

		// Get info
		r, err := c.Enumerate(context.Background(), config.req)
		assert.NoError(t, err)
		assert.Len(t, r.Alerts, config.expected)
	}
}

// TestAlertsServerEnumerateError tests errors returned from server code.
func TestAlertsServerEnumerateError(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageAlertsClient(s.Conn())

	errs := []error{
		status.Error(codes.InvalidArgument, "must provide a query"),
		status.Error(codes.Internal, errors.New("kvdb error").Error()),
		status.Error(codes.DeadlineExceeded,
			"deadline is reached, server side func exiting"),
	}

	myAlerts := make([]*api.Alert, 0)

	for _, err := range errs {
		var filters []interface{}
		for _, filter := range getFilters(configs[0].req.Queries) {
			filters = append(filters, filter)
		}

		s.MockFilterDeleter().EXPECT().Enumerate(filters...).Return(myAlerts, err).Times(1)
		// Get info
		_, outErr := c.Enumerate(context.Background(), configs[0].req)
		assert.Error(t, outErr, err.Error())
	}
}

// TestAlertsServerDelete tests delete functionality over gRPC using mock.
func TestAlertsServerDelete(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageAlertsClient(s.Conn())

	for _, config := range configs {
		var filters []interface{}
		for _, filter := range getFilters(config.req.Queries) {
			filters = append(filters, filter)
		}
		myAlerts := make([]*api.Alert, config.expected)
		for i := range myAlerts {
			myAlerts[i] = new(api.Alert)
		}
		s.MockFilterDeleter().EXPECT().Delete(filters...).Return(nil).Times(1)

		// Get info
		_, err := c.Delete(context.Background(), &api.SdkAlertsDeleteRequest{
			Queries: config.req.Queries,
		})
		assert.NoError(t, err)
	}
}

// TestAlertsServerDeleteError tests errors returned from server code.
func TestAlertsServerDeleteError(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageAlertsClient(s.Conn())

	errs := []error{
		status.Error(codes.InvalidArgument, "must provide a query"),
		status.Error(codes.Internal, errors.New("kvdb error").Error()),
		status.Error(codes.DeadlineExceeded,
			"deadline is reached, server side func exiting"),
	}

	for _, err := range errs {
		var filters []interface{}
		for _, filter := range getFilters(configs[0].req.Queries) {
			filters = append(filters, filter)
		}

		s.MockFilterDeleter().EXPECT().Delete(filters...).Return(err).Times(1)
		// Get info
		_, outErr := c.Delete(context.Background(), &api.SdkAlertsDeleteRequest{
			Queries: configs[0].req.Queries,
		})
		assert.Error(t, outErr, err.Error())
	}
}
