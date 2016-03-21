package flexvolume

import (
	"fmt"
	"os"

	"go.pedge.io/dlog"
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
)

type client struct {
	apiClient APIClient
}

const (
	volumeIDKey = "volumeID"
)

var (
	successBytes = []byte(`{"Status":"Success"}`)
)

func newClient(apiClient APIClient) *client {
	return &client{apiClient}
}

func (c *client) Init() error {
	_, err := c.apiClient.Init(
		context.Background(),
		google_protobuf.EmptyInstance,
	)
	return err
}

func (c *client) Attach(jsonOptions map[string]string) error {
	var output []byte

	_, err := c.apiClient.Attach(
		context.Background(),
		&AttachRequest{
			JsonOptions: jsonOptions,
		},
	)
	if err == nil {
		output = []byte(fmt.Sprintf(`{"Status":"Success", "Device":"%s"}`, jsonOptions[volumeIDKey]))
	} else {
		output = newFailureBytes(err)
	}
	if _, osErr := os.Stdout.Write(output); osErr != nil {
		dlog.Warnf("Unable to write output to stdout : %s", osErr.Error())
	}
	return err
}

func (c *client) Detach(mountDevice string) error {
	_, err := c.apiClient.Detach(
		context.Background(),
		&DetachRequest{
			MountDevice: mountDevice,
		},
	)

	output := newOutput(err)
	if _, osErr := os.Stdout.Write(output); osErr != nil {
		dlog.Warnf("Unable to write output to stdout : %s", osErr.Error())
	}

	return err
}

func (c *client) Mount(targetMountDir string, mountDevice string, jsonOptions map[string]string) error {
	var err error
	if err = os.MkdirAll(targetMountDir, os.ModeDir); err == nil {
		_, err = c.apiClient.Mount(
			context.Background(),
			&MountRequest{
				TargetMountDir: targetMountDir,
				MountDevice:    mountDevice,
				JsonOptions:    jsonOptions,
			},
		)
	}
	output := newOutput(err)

	if _, osErr := os.Stdout.Write(output); osErr != nil {
		dlog.Warnf("Unable to write output to stdout : %s", osErr.Error())
	}

	return err
}

func (c *client) Unmount(mountDir string) error {
	_, err := c.apiClient.Unmount(
		context.Background(),
		&UnmountRequest{
			MountDir: mountDir,
		},
	)
	output := newOutput(err)

	if _, osErr := os.Stdout.Write(output); osErr != nil {
		dlog.Warnf("Unable to write output to stdout : %s", osErr.Error())
	}

	return err
}

func newFailureBytes(err error) []byte {
	return []byte(fmt.Sprintf(`{"Status":"Failure", "Message":"%s"}`, err.Error()))
}

func newOutput(err error) []byte {
	if err != nil {
		return newFailureBytes(err)
	} else {
		return successBytes
	}
}
