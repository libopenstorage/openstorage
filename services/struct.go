package services

// swagger:model
type ServiceMessage struct {
	Status  string
	Version string
}

// AddDrive specifies parameter for drive add service operation
// swagger:model
type AddDrive struct {
	Operation string
	Drive     string
	Journal   bool
}

// ReplaceDrive specifies Source and Target parametes for replace drive operation
// swagger:model
type ReplaceDrive struct {
	Operation string
	Source    string
	Target    string
}

// RebalancePool specifies PoolID to perform rebalance operation
// swagger:model
type RebalancePool struct {
	Operation string
	PoolID    int
}
