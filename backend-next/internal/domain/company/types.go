package company

const (
	ArrivalStoryID        = "chapter1Arrival"
	ArrivalStoryFirstStep = "chapter-title"
)

// StoryProgress records the server-authoritative position of a story.
type StoryProgress struct {
	Status string `json:"status"`
	StepID string `json:"stepId"`
}

// Company represents a player's company.
type Company struct {
	ID             int            `json:"id"`
	PlayerID       int            `json:"player_id"`
	Name           string         `json:"name"`
	Money          float64        `json:"money"`
	Level          int            `json:"level"`
	XP             int64          `json:"xp"`
	Preferences    map[string]any `json:"preferences,omitempty"` // dynamic keys (story IDs); structure is too dynamic for a fixed struct, safe to remain map[string]any
	Buildings      []Building     `json:"buildings,omitempty"`
	Inventory      map[int]int    `json:"inventory,omitempty"`
	Executives     []Executive    `json:"executives,omitempty"`
	WarehouseLevel int            `json:"warehouse_level"`
	CreatedAt      string         `json:"created_at"`
	LastRetailAt   string         `json:"last_retail_at,omitempty"`
}

// NewPlayerPreferences returns preferences for a newly registered account.
func NewPlayerPreferences() map[string]any {
	return map[string]any{
		"storyProgress": map[string]any{
			ArrivalStoryID: map[string]any{
				"status": "not_started",
				"stepId": ArrivalStoryFirstStep,
			},
		},
	}
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

// Executive represents a hired executive for a company.
type Executive struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Title           string  `json:"title"`
	Level           int     `json:"level"`
	Rarity          string  `json:"rarity"`
	Stage           string  `json:"stage"`
	Salary          float64 `json:"salary"`
	ProductionBonus float64 `json:"productionBonus"`
	SalesBonus      float64 `json:"salesBonus"`
	MgmtDiscount    float64 `json:"mgmtDiscount"`
	Morale          int     `json:"morale,omitempty"`
}

// CompanyResponse is the public-facing company DTO.
type CompanyResponse struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Money      float64     `json:"money"`
	Level      int         `json:"level"`
	XP         int64       `json:"xp"`
	Buildings  []Building  `json:"buildings"`
	Inventory  map[int]int `json:"inventory"`
	Executives []Executive `json:"executives,omitempty"`
}

// ToResponse converts a Company to its API response form.
func (c *Company) ToResponse() CompanyResponse {
	return CompanyResponse{
		ID:         c.ID,
		Name:       c.Name,
		Money:      c.Money,
		Level:      c.Level,
		XP:         c.XP,
		Buildings:  c.Buildings,
		Inventory:  c.Inventory,
		Executives: c.Executives,
	}
}
