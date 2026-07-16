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
	// RetailCarry preserves fractional shelf demand between short player
	// catch-ups. Without it, routine profile polling would round every small
	// interval down to zero and low-volume stores would never make a sale.
	RetailCarry map[string]float64 `json:"retail_carry,omitempty"`
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
	ID                    string      `json:"id"`
	BuildingID            int         `json:"building_id"`
	Kind                  int         `json:"kind"`
	Name                  string      `json:"name"`
	Level                 int         `json:"level"`
	MapID                 string      `json:"map_id"`
	SlotID                string      `json:"slot_id"`
	X                     int         `json:"x"`
	Y                     int         `json:"y"`
	RobotCount            int         `json:"robot_count"`
	Shelves               []ShelfItem `json:"shelves,omitempty"`
	PurchaseRequestID     string      `json:"purchase_request_id,omitempty"`
	PurchaseCatalogItemID string      `json:"purchase_catalog_item_id,omitempty"`
	PurchaseCost          float64     `json:"purchase_cost,omitempty"`
	// UpgradeTargetLevel and its timestamps describe an in-progress upgrade.
	// The effective Level changes only once the construction timer completes.
	UpgradeTargetLevel int    `json:"upgrade_target_level,omitempty"`
	UpgradeStartedAt   string `json:"upgrade_started_at,omitempty"`
	UpgradeCompletesAt string `json:"upgrade_completes_at,omitempty"`
}

// ShelfItem represents items stocked in a retail building's shelf.
type ShelfItem struct {
	ResourceID int     `json:"resource_id"`
	Quality    int     `json:"quality"`
	Quantity   int     `json:"quantity"`
	MaxQty     int     `json:"max_qty"`
	Price      float64 `json:"price"`
	PriceLock  bool    `json:"price_lock"`
	Revenue    float64 `json:"revenue,omitempty"`
}

// ExecutivePosition is one of the four company leadership assignments. An
// executive may be hired without an assignment, but only an assigned executive
// contributes their specialised skill to the live game loop.
type ExecutivePosition string

const (
	ExecutivePositionUnassigned ExecutivePosition = ""
	ExecutivePositionCOO        ExecutivePosition = "coo"
	ExecutivePositionCFO        ExecutivePosition = "cfo"
	ExecutivePositionCMO        ExecutivePosition = "cmo"
	ExecutivePositionCTO        ExecutivePosition = "cto"
)

// ExecutiveSkills deliberately uses legible, broad skills instead of a hidden
// per-level bonus table. A candidate's specialty is weighted toward the skill
// used by its main position.
type ExecutiveSkills struct {
	Management    float64 `json:"management"`
	Accounting    float64 `json:"accounting"`
	Communication float64 `json:"communication"`
	Science       float64 `json:"science"`
}

// SkillForPosition returns the skill a leadership position evaluates.
func (s ExecutiveSkills) SkillForPosition(position ExecutivePosition) float64 {
	switch position {
	case ExecutivePositionCOO:
		return s.Management
	case ExecutivePositionCFO:
		return s.Accounting
	case ExecutivePositionCMO:
		return s.Communication
	case ExecutivePositionCTO:
		return s.Science
	default:
		return 0
	}
}

// EffectiveExecutiveSkill applies the published diminishing-return curve to
// an assigned position's raw skill. It keeps a very strong executive useful
// without letting a single number scale linearly forever.
func EffectiveExecutiveSkill(raw float64) float64 {
	if raw <= 0 {
		return 0
	}
	switch {
	case raw <= 60:
		return raw
	case raw <= 80:
		return 60 + (raw-60)/2
	default:
		return 70 + (raw-80)/2
	}
}

// ActiveExecutiveSkill returns the effective skill of the executive assigned
// to a position. Legacy executives without position data deliberately retain
// no new hidden effects; players can inspect and assign them explicitly.
func ActiveExecutiveSkill(executives []Executive, position ExecutivePosition) float64 {
	for _, executive := range executives {
		if executive.Position == position {
			return EffectiveExecutiveSkill(executive.Skills.SkillForPosition(position))
		}
	}
	return 0
}

// Executive represents a hired executive for a company.
type Executive struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Title           string            `json:"title"`
	Specialty       ExecutivePosition `json:"specialty,omitempty"`
	Position        ExecutivePosition `json:"position,omitempty"`
	Skills          ExecutiveSkills   `json:"skills"`
	Level           int               `json:"level"`
	Rarity          string            `json:"rarity"`
	Stage           string            `json:"stage"`
	Salary          float64           `json:"salary"`
	ProductionBonus float64           `json:"productionBonus"`
	SalesBonus      float64           `json:"salesBonus"`
	MgmtDiscount    float64           `json:"mgmtDiscount"`
	Morale          int               `json:"morale,omitempty"`
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
