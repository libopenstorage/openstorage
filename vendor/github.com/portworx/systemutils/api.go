package systemutils

// Lun describes the properties of a physical LUN.
type Lun struct {
	Iops      int
	Capacity  uint64
	Available uint64
}

type System interface {
	// CpuUsage returns the usage in percentage format.
	CpuUsage() (usage float64, total float64, ticks float64)

	// MemUsage returns the available memory on this system.
	MemUsage() (available float64)

	// Luns returns information on the available LUNs on this system.
	Luns() (luns map[string]Lun)
}

func New() System {
	return system{}
}
