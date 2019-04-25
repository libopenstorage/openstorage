/*
Copyright 2019 Portworx

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
	"fmt"
	grpcstatus "google.golang.org/grpc/status"
)

func (s *SdkErrorMessage) GetGrpcStatus() *grpcstatus.Status {
	if s != nil {
		return grpcstatus.FromProto(s.GetStatus())
	}
	return nil
}

func (s *SdkErrorMessage) Stack() string {
	if s == nil {
		return ""
	}

	e := fmt.Sprintf("%s\n", s.Error())
	for _, detail := range s.Details() {
		e += fmt.Sprintf("  %s\n", detail.Error())
	}
	return e
}

func (s *SdkErrorMessage) Details() []*SdkErrorMessage {
	if s == nil {
		return nil
	}

	details := s.GetGrpcStatus().Details()
	messages := make([]*SdkErrorMessage, 0, len(details))
	for _, detail := range details {
		switch t := detail.(type) {
		case *SdkErrorMessage:
			messages = append(messages, t)
		}
	}
	return messages
}

func (s *SdkErrorMessage) Error() string {
	if s != nil {
		return s.GetGrpcStatus().Err().Error()
	}
	return ""
}
