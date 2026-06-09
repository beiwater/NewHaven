package system

// DevBootstrapConfig holds the configuration for development auto-setup.
type DevBootstrapConfig struct {
	Username          string
	Password          string
	StartingMoney     float64
	StartingLevel     int
	StartingBuildings []int // building catalog IDs to auto-create
}
