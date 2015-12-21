package volume

type IoNotSupported struct {}

func (io *IoNotSupported) Read(
	volumeID string,
	buffer []byte,
	size uint64,
	offset int64,
) (int64, error) {
	return 0, ErrNotSupported
}

func (io *IoNotSupported) Write(
	volumeID string,
	buffer []byte,
	size uint64,
	offset int64,
) (int64, error) {
	return 0, ErrNotSupported
}

func (io *IoNotSupported) Flush(volumeID string) error {
	return ErrNotSupported
}
