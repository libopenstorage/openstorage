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
	"time"

	"github.com/libopenstorage/openstorage/pkg/sched"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gopkg.in/yaml.v2"
)

const (
	// Max day for each month. All months have at least 28 days
	maxDay = int32(28)
	// Max hour
	maxHour = int32(23)
	// Max minute
	maxMinute = int32(59)
)

func sdkWeekdayToTimeWeekday(weekday api.SdkTimeWeekday) time.Weekday {
	// Purposely not using math to translate in case the values are ever changed
	switch weekday {
	case api.SdkTimeWeekday_SdkTimeWeekdaySunday:
		return time.Sunday
	case api.SdkTimeWeekday_SdkTimeWeekdayMonday:
		return time.Monday
	case api.SdkTimeWeekday_SdkTimeWeekdayTuesday:
		return time.Tuesday
	case api.SdkTimeWeekday_SdkTimeWeekdayWednesday:
		return time.Wednesday
	case api.SdkTimeWeekday_SdkTimeWeekdayThursday:
		return time.Thursday
	case api.SdkTimeWeekday_SdkTimeWeekdayFriday:
		return time.Friday
	case api.SdkTimeWeekday_SdkTimeWeekdaySaturday:
		return time.Saturday
	}
	panic("Illegal time of the week")
}

func timeWeekdayToSdkWeekly(t time.Weekday) api.SdkTimeWeekday {
	// Purposely not using math to translate in case the values are ever changed
	switch t {
	case time.Sunday:
		return api.SdkTimeWeekday_SdkTimeWeekdaySunday
	case time.Monday:
		return api.SdkTimeWeekday_SdkTimeWeekdayMonday
	case time.Tuesday:
		return api.SdkTimeWeekday_SdkTimeWeekdayTuesday
	case time.Wednesday:
		return api.SdkTimeWeekday_SdkTimeWeekdayWednesday
	case time.Thursday:
		return api.SdkTimeWeekday_SdkTimeWeekdayThursday
	case time.Friday:
		return api.SdkTimeWeekday_SdkTimeWeekdayFriday
	case time.Saturday:
		return api.SdkTimeWeekday_SdkTimeWeekdaySaturday
	}
	panic("Illegal time of the week")
}

func sdkSchedToRetainInternalSpec(
	req *api.SdkSchedulePolicyInterval,
) (*sched.RetainIntervalSpec, error) {

	if req.GetRetain() < 1 {
		return nil, status.Error(codes.InvalidArgument, "Must retain more than 0")
	}

	// Translate sdk schedule to yaml RetainIntervalSpec string.
	var spec sched.IntervalSpec
	if daily := req.GetDaily(); daily != nil {
		// daily
		if daily.GetHour() < 0 || daily.GetHour() > maxDay {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid hour value: %d", daily.GetHour())
		} else if daily.GetMinute() < 0 || daily.GetMinute() > maxMinute {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid minute value: %d", daily.GetMinute())
		}
		spec = sched.Daily(
			int(daily.GetHour()),
			int(daily.GetMinute())).
			Spec()
	} else if weekly := req.GetWeekly(); weekly != nil {
		// weekly
		if weekly.GetDay() < api.SdkTimeWeekday_SdkTimeWeekdaySunday ||
			weekly.GetDay() > api.SdkTimeWeekday_SdkTimeWeekdaySaturday {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid weekday value: %d", weekly.GetDay())
		} else if weekly.GetHour() < 0 || weekly.GetHour() > maxDay {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid hour value: %d", weekly.GetHour())
		} else if weekly.GetMinute() < 0 || weekly.GetMinute() > maxMinute {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid minute value: %d", weekly.GetMinute())
		}
		spec = sched.Weekly(
			sdkWeekdayToTimeWeekday(weekly.GetDay()),
			int(daily.GetHour()),
			int(daily.GetMinute())).
			Spec()
	} else if monthly := req.GetMonthly(); monthly != nil {
		// monthly
		if monthly.GetDay() < 1 || monthly.GetDay() > maxDay {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid day value: %d", monthly.GetDay())
		} else if monthly.GetHour() < 0 || monthly.GetHour() > maxDay {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid hour value: %d", monthly.GetHour())
		} else if monthly.GetMinute() < 0 || monthly.GetMinute() > maxMinute {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid minute value: %d", monthly.GetMinute())
		}
		spec = sched.Monthly(
			int(monthly.GetDay()),
			int(monthly.GetHour()),
			int(monthly.GetMinute())).
			Spec()
	} else {
		return nil, status.Error(codes.InvalidArgument, "Invalid schedule period type")
	}

	return &sched.RetainIntervalSpec{
		IntervalSpec: spec,
		Retain:       uint32(req.GetRetain()),
	}, nil
}

func sdkSchedToRetainInternalSpecYamlByte(req *api.SdkSchedulePolicyInterval) ([]byte, error) {
	sched, err := sdkSchedToRetainInternalSpec(req)
	if err != nil {
		return nil, err
	}

	out, err := yaml.Marshal(sched)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create schedule: %v", err)
	}

	return out, nil
}

func retainInternalSpecToSdkSched(spec *sched.RetainIntervalSpec) (*api.SdkSchedulePolicyInterval, error) {

	var resp *api.SdkSchedulePolicyInterval
	switch spec.Freq {
	case sched.MonthlyType:
		resp = &api.SdkSchedulePolicyInterval{
			PeriodType: &api.SdkSchedulePolicyInterval_Monthly{
				Monthly: &api.SdkSchedulePolicyIntervalMonthly{
					Day:    int32(spec.Day),
					Hour:   int32(spec.Hour),
					Minute: int32(spec.Minute),
				},
			},
		}
	case sched.WeeklyType:
		resp = &api.SdkSchedulePolicyInterval{
			PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
				Weekly: &api.SdkSchedulePolicyIntervalWeekly{
					Day:    timeWeekdayToSdkWeekly(time.Weekday(spec.Weekday)),
					Hour:   int32(spec.Hour),
					Minute: int32(spec.Minute),
				},
			},
		}
	case sched.DailyType:
		resp = &api.SdkSchedulePolicyInterval{
			PeriodType: &api.SdkSchedulePolicyInterval_Daily{
				Daily: &api.SdkSchedulePolicyIntervalDaily{
					Hour:   int32(spec.Hour),
					Minute: int32(spec.Minute),
				},
			},
		}
	default:
		return nil, status.Errorf(codes.Internal, "Unknown schedule type: %s", spec.Freq)
	}

	resp.Retain = int64(spec.Retain)
	return resp, nil
}

func retainInternalSpecYamlByteToSdkSched(
	in []byte,
) (*api.SdkSchedulePolicyInterval, error) {

	// Get spec from yaml
	var spec sched.RetainIntervalSpec
	err := yaml.Unmarshal(in, &spec)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to retreive schedule")
	}

	return retainInternalSpecToSdkSched(&spec)
}
