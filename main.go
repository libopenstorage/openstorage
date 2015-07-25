package main

import (
	"github.com/libopenstorage/ec2driver"
	"github.com/libopenstorage/volume"
	"os"
)

func main() {
	v := volume.NewVolumePlugin()
	v.Listen(os.Args[1])
}
