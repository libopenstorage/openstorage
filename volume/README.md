## Volume Drivers

Volume drivers implement the [Volume Plugin Interface](https://docs.docker.com/engine/extend/plugins_volume/).
This provides an interface to register a volume driver and advertise the driver to Docker.  Registering a driver with this volume interface will cause Docker to be able to communicate with the driver to create and assign volumes to a container.

### Block Drivers
Block drivers operate at the block layer.  They provide raw volumes formatted with a user specified filesystem.  This volume is then mounted into the container at a path specified using the `docker run -v` option.

### File Drivers
File drivers operate at the filesystem layer.
