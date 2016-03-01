package flexvolume

import (
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
)

type client struct {
	apiClient APIClient
}

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

func (c *client) Attach(jsonOptions map[string]interface{}) error {
	value, err := JSONOptionsToBytes(jsonOptions)
	if err != nil {
		return err
	}
	_, err = c.apiClient.Attach(
		context.Background(),
		&AttachRequest{
			JsonOptions: value,
		},
	)
	return err
}

func (c *client) Detach(mountDevice string) error {
	_, err := c.apiClient.Detach(
		context.Background(),
		&DetachRequest{
			MountDevice: mountDevice,
		},
	)
	return err
}

func (c *client) Mount(targetMountDir string, mountDevice string, jsonOptions map[string]interface{}) error {
	value, err := JSONOptionsToBytes(jsonOptions)
	if err != nil {
		return err
	}
	_, err = c.apiClient.Mount(
		context.Background(),
		&MountRequest{
			TargetMountDir: targetMountDir,
			MountDevice:    mountDevice,
			JsonOptions:    value,
		},
	)
	return err
}

func (c *client) Unmount(mountDir string) error {
	_, err := c.apiClient.Unmount(
		context.Background(),
		&UnmountRequest{
			MountDir: mountDir,
		},
	)
	return err
}
