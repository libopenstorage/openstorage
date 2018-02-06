/*
Copyright 2017 Luis Pabón luis@portworx.com

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
	"regexp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	context "golang.org/x/net/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	csiClientVersion = &csi.Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
	}
)

var _ = Describe("GetSupportedVersions [Identity Server]", func() {
	var (
		c csi.IdentityClient
	)

	BeforeEach(func() {
		c = csi.NewIdentityClient(conn)
	})

	It("should return an array of supported versions", func() {
		res, err := c.GetSupportedVersions(
			context.Background(),
			&csi.GetSupportedVersionsRequest{})

		By("checking response to have supported versions list")
		Expect(err).NotTo(HaveOccurred())
		Expect(res.GetSupportedVersions()).NotTo(BeNil())
		Expect(len(res.GetSupportedVersions()) >= 1).To(BeTrue())

		By("checking each version")
		for _, version := range res.GetSupportedVersions() {
			Expect(version).NotTo(BeNil())
			Expect(version.GetMajor()).To(BeNumerically("<", 100))
			Expect(version.GetMinor()).To(BeNumerically("<", 100))
			Expect(version.GetPatch()).To(BeNumerically("<", 100))
		}
	})
})

var _ = Describe("GetPluginInfo [Identity Server]", func() {
	var (
		c csi.IdentityClient
	)

	BeforeEach(func() {
		c = csi.NewIdentityClient(conn)
	})

	It("should fail when no version is provided", func() {
		_, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
		Expect(err).To(HaveOccurred())

		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(Equal(codes.InvalidArgument))
	})

	It("should return appropriate information", func() {
		req := &csi.GetPluginInfoRequest{
			Version: csiClientVersion,
		}
		res, err := c.GetPluginInfo(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())

		By("verifying name size and characters")
		Expect(res.GetName()).ToNot(HaveLen(0))
		Expect(len(res.GetName())).To(BeNumerically("<=", 63))
		Expect(regexp.
			MustCompile("^[a-zA-Z][A-Za-z0-9-\\.\\_]{0,61}[a-zA-Z]$").
			MatchString(res.GetName())).To(BeTrue())
	})
})
