package company

// Company represents a player's company.
type Company struct {
	ID           int            `json:"id"`
	PlayerID     int            `json:"player_id"`
	Name         string         `json:"name"`
	Money        float64        `json:"money"`
	Level        int            `json:"level"`
	XP           int64          `json:"xp"`
	Preferences  map[string]any `json:"preferences,omitempty"`
	Buildings    []Building     `json:"buildings,omitempty"`
	Inventory    map[int]int    `json:"inventory,omitempty"`
	CreatedAt    string         `json:"created_at"`
}

// Building represents a building placed on a map.
type Building struct {
	ID         string `json:"id"`
	BuildingID int    `json:"building_id"`
	Kind       int    `json:"kind"`
	Name       string `json:"name"`
	Level      int    `json:"level"`
	MapID      string `json:"map_id"`
	SlotID     string `json:"slot_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	RobotCount int    `json:"robot_count"`
}

// CompanyResponse is the public-facing company DTO.
type CompanyResponse struct {
	ID        int              `json:"id"`
	Name      string           `json:"name"`
	Money     float64          `json:"money"`
	Level     int              `json:"level"`
	XP        int64            `json:"xp"`
	Buildings []Building       `json:"buildings"`
	Inventory map[int]int      `json:"inventory"`
}

// ToResponse converts a Company to its API response form.
func (c *Company) ToResponse() CompanyResponse {
	return CompanyResponse{
		ID:        c.ID,
		Name:      c.Name,
		Money:     c.Money,
		Level:     c.Level,
		XP:        c.XP,
		Buildings: c.Buildings,
		Inventory: c.Inventory,
	}
}
