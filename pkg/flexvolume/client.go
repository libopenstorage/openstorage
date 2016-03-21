package flexvolume

import (
	"encoding/json"
	"os"

	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
)

type client struct {
	apiClient APIClient
}

type AttachSuccessOutput struct {
	Status string
	Device string
}

type SuccessOutput struct {
	Status string
}

type FailureOutput struct {
	Status  string
	Message string
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

func (c *client) Attach(jsonOptions map[string]string) error {
	var output []byte
	_, err := c.apiClient.Attach(
		context.Background(),
		&AttachRequest{
			JsonOptions: jsonOptions,
		},
	)
	if err == nil {
		a := AttachSuccessOutput{
			Status: "Success",
			Device: jsonOptions["volumeID"],
		}
		output, _ = json.Marshal(a)
	} else {
		a := FailureOutput{
			Status:  "Failure",
			Message: err.Error(),
		}
		output, _ = json.Marshal(a)
	}

	os.Stdout.Write(output)
	return err
}

func (c *client) Detach(mountDevice string) error {
	var output []byte

	_, err := c.apiClient.Detach(
		context.Background(),
		&DetachRequest{
			MountDevice: mountDevice,
		},
	)

	if err == nil {
		a := SuccessOutput{
			Status: "Success",
		}
		output, _ = json.Marshal(a)
	} else {
		a := FailureOutput{
			Status:  "Failure",
			Message: err.Error(),
		}
		output, _ = json.Marshal(a)
	}

	os.Stdout.Write(output)
	return err
}

func (c *client) Mount(targetMountDir string, mountDevice string, jsonOptions map[string]string) error {
	err := os.MkdirAll(targetMountDir, os.ModeDir)
	var output []byte

	if err == nil {
		_, err = c.apiClient.Mount(
			context.Background(),
			&MountRequest{
				TargetMountDir: targetMountDir,
				MountDevice:    mountDevice,
				JsonOptions:    jsonOptions,
			},
		)
		if err == nil {
			a := SuccessOutput{
				Status: "Success",
			}
			output, _ = json.Marshal(a)
		} else {
			a := FailureOutput{
				Status:  "Failure",
				Message: err.Error(),
			}
			output, _ = json.Marshal(a)
		}
	} else {
		a := FailureOutput{
			Status:  "Failure",
			Message: err.Error(),
		}
		output, _ = json.Marshal(a)
	}

	os.Stdout.Write(output)
	return err
}

func (c *client) Unmount(mountDir string) error {
	var output []byte

	_, err := c.apiClient.Unmount(
		context.Background(),
		&UnmountRequest{
			MountDir: mountDir,
		},
	)

	if err == nil {
		a := SuccessOutput{
			Status: "Success",
		}
		output, _ = json.Marshal(a)
	} else {
		a := FailureOutput{
			Status:  "Failure",
			Message: err.Error(),
		}
		output, _ = json.Marshal(a)
	}

	os.Stdout.Write(output)

	return err
}
