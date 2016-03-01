package server

type flexVolumeClient struct{}

func newFlexVolumeClient() *flexVolumeClient {
	return &flexVolumeClient{}
}

func (c *flexVolumeClient) Init() error {
	return nil
}

func (c *flexVolumeClient) Attach(jsonOptions map[string]string) error {
	return nil
}

func (c *flexVolumeClient) Detach(mountDevice string) error {
	return nil
}

func (c *flexVolumeClient) Mount(targetMountDir string, mountDevice string, jsonOptions map[string]string) error {
	return nil
}

func (c *flexVolumeClient) Unmount(mountDir string) error {
	return nil
}
