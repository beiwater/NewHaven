package model

type ResourceAmount struct {
	ResourceID int `json:"resourceId"`
	Quantity   int `json:"quantity"`
}

// Bond represents a bond in the game
type Bond struct {
	ID                  string  `json:"id"`
	Amount              int     `json:"amount"`
	Interest            float64 `json:"interest"`
	PurchasedAt         string  `json:"purchased_at"`
	MissedPayments      int     `json:"missed_payments"`
	InterestCollected   float64 `json:"interestCollected"`
	RatingWhenPurchased string  `json:"ratingWhenPurchased"`
	IssuerCompanyID     int     `json:"issuerCompanyId"`
	OwnerCompanyID      int     `json:"ownerCompanyId"`
	CallableAfter       string  `json:"callableAfter"`
	RestructurePct      float64 `json:"restructure_percentage"`
}

// GovContract represents a government procurement contract
type GovContract struct {
	ID              string           `json:"id"`
	ResourceID      int              `json:"resourceId"`
	Quality         int              `json:"quality"`
	Quantity        int              `json:"quantity"`
	MaxPrice        float64          `json:"maxPrice"`
	DepositRate     float64          `json:"depositRate"`
	Status          string           `json:"status"`
	Bids            []map[string]any `json:"bids"`
	WinnerCompanyID int              `json:"winnerCompanyId"`
	WinningPrice    float64          `json:"winningPrice,omitempty"`
	AwardedAt       string           `json:"awardedAt,omitempty"`
	DueAt           string           `json:"dueAt,omitempty"`
	DeliveredAt     string           `json:"deliveredAt,omitempty"`
	DefaultedAt     string           `json:"defaultedAt,omitempty"`
	Penalty         string           `json:"penalty,omitempty"`
}

// LedgerEntry represents a single financial ledger entry
type LedgerEntry struct {
	ID        string         `json:"id"`
	At        string         `json:"at"`
	Kind      string         `json:"kind"`
	Amount    float64        `json:"amount"`
	Direction string         `json:"direction"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// Notification represents a game notification
type Notification struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Read  bool   `json:"read"`
}

// Message represents a chat message
type Message struct {
	ID       string `json:"id"`
	Chatroom string `json:"chatroom"`
	Body     string `json:"body"`
	At       string `json:"at"`
	From     string `json:"from,omitempty"`
	FromID   int    `json:"fromId,omitempty"`
	Token    string `json:"token,omitempty"`
}

type MarketOrder struct {
	ID         string  `json:"id"`
	ResourceID int     `json:"resourceId"`
	Kind       int     `json:"kind"` // 0=sell,1=buy
	Price      float64 `json:"price"`
	Quality    int     `json:"quality"`
	Quantity   int     `json:"quantity"`
	Remaining  int     `json:"remaining"`
	CompanyID  int     `json:"companyId"`
	CreatedAt  string  `json:"createdAt"`
	Status     string  `json:"status"` // open, filled, cancelled
}

type Trade struct {
	ID          string  `json:"id"`
	ResourceID  int     `json:"resourceId"`
	Quality     int     `json:"quality"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	BuyOrderID  string  `json:"buyOrderId"`
	SellOrderID string  `json:"sellOrderId"`
	CreatedAt   string  `json:"createdAt"`
}

type Company struct {
	ID                int              `json:"id"`
	Name              string           `json:"name"`
	Money             float64          `json:"money"`
	Level             int              `json:"level"`
	ProductionSlots   int              `json:"productionSlots"`
	Inventory         map[int]int      `json:"inventory"`
	QualityInventory  map[string]int   `json:"qualityInventory,omitempty"`
	PlacedBuildings   []map[string]any `json:"placedBuildings"`
	UnplacedBuildings []map[string]any `json:"unplacedBuildings"`
	WarehouseLevel    int              `json:"warehouseLevel"`
}

type Player struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Token        string `json:"token,omitempty"`
	CompanyID    int    `json:"companyId"`
	RegisteredAt string `json:"registeredAt"`
}

type ProductionJob struct {
	ID              string         `json:"id"`
	BuildingID      string         `json:"buildingId"`
	ResourceID      int            `json:"resourceId"`
	Amount          int            `json:"amount"`
	Quality         int            `json:"quality"`
	Input           map[int]int    `json:"input"`
	Output          map[int]int    `json:"output"`
	ClaimedAmount   int            `json:"claimedAmount"`
	ClaimableAmount int            `json:"claimableAmount,omitempty"`
	XPAwarded       int            `json:"xpAwarded,omitempty"`
	StartedAt       string         `json:"startedAt"`
	CompletesAt     string         `json:"completesAt"`
	Status          string         `json:"status"`
	Meta            map[string]any `json:"meta,omitempty"`
}

type Auction struct {
	ID            string       `json:"id"`
	Item          string       `json:"item"`
	ItemID        string       `json:"itemId"`
	StartingBid   float64      `json:"startingBid"`
	CurrentBid    float64      `json:"currentBid"`
	HighestBidder int          `json:"highestBidder"`
	EndsAt        string       `json:"endsAt"`
	Status        string       `json:"status"` // open|awarded
	Bids          []AuctionBid `json:"bids"`
	CreatedAt     string       `json:"createdAt"`
}

type Order struct {
	ID          string  `json:"id"`
	ResourceID  int     `json:"resourceId"`
	Quality     int     `json:"quality"`
	Quantity    int     `json:"quantity"`
	RewardCash  float64 `json:"rewardCash"`
	RewardXP    int     `json:"rewardXP"`
	Status      string  `json:"status"` // active | completed | expired
	CreatedAt   string  `json:"createdAt"`
	CompletedAt string  `json:"completedAt,omitempty"`
}

type AuctionBid struct {
	CompanyID int     `json:"companyId"`
	Amount    float64 `json:"amount"`
	At        string  `json:"at"`
}

type ResearchProject struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Building          string      `json:"building"`
	ResourceCost      map[int]int `json:"resourceCost"`
	CashCost          float64     `json:"cashCost"`
	DurationHours     int         `json:"durationHours"`
	UnlockRecipeID    int         `json:"unlockRecipeId"`    // resource ID this unlocks
	QualityResourceID int         `json:"qualityResourceId"` // resource ID whose quality improves on completion
	UnlockPct         float64     `json:"unlockPct"`         // efficiency gain %
	Status            string      `json:"status"`            // available|in_progress|completed
	Progress          int         `json:"progress"`          // 0-100
	StartedAt         string      `json:"startedAt,omitempty"`
	CompletesAt       string      `json:"completesAt,omitempty"`
}

type GameState struct {
	Players             []Player
	NextPlayerID        int
	Companies           []Company
	Orders              []MarketOrder
	Trades              []Trade
	CSRFToken           string
	Messages            []Message
	Bonds               []Bond
	ContractsIn         []map[string]any
	ContractsOut        []map[string]any
	Achievements        []map[string]any
	Notifications       []Notification
	PlayerPreferences   map[string]any
	BotCompanies        []Company
	MarketTicks         map[int][]map[string]any
	ProductionJobs      []ProductionJob
	GovernmentContracts []GovContract
	Executives          []map[string]any
	Ledger              []LedgerEntry
	XP                  int
	XpToNextLevel       int
	MarketPressure      map[int]float64 // per-resource buy/sell pressure (-1..1)
	UnlockedRecipes     map[int]bool
	ResearchedQuality   map[int]int `json:"researchedQuality"` // resourceID -> quality level
	Auctions            []Auction
	DailyOrders         []Order `json:"dailyOrders"`
	DailyOrdersDate     string  `json:"dailyOrdersDate"`
	LastActiveAt        string  `json:"lastActiveAt"`
	BoostType           string  `json:"boostType"`
	SimulatedAt         string  `json:"simulatedAt"` // empty = real time, set = frozen time for testing
	BoostEndsAt         string  `json:"boostEndsAt"`
	BoostMultiplier     float64 `json:"boostMultiplier"`
	ResearchProjects    []ResearchProject
	ProcessedRequests   map[string]map[string]any `json:"-"` // requestId → cached response, not persisted
	// Market competition / national team fields
	DailyTradeVolume   map[int]float64 `json:"dailyTradeVolume"`
	DailyTradeQty      map[int]int     `json:"dailyTradeQty"`
	DailyHighPrice     map[int]float64 `json:"dailyHighPrice"`
	DailyLowPrice      map[int]float64 `json:"dailyLowPrice"`
	YesterdayVolume    map[int]float64 `json:"yesterdayVolume"`
	YesterdayHighPrice map[int]float64 `json:"yesterdayHighPrice"`
	YesterdayClose     map[int]float64 `json:"yesterdayClose"`
	MarketLocked       map[int]bool    `json:"marketLocked"`
	NationalTeamActive map[int]bool    `json:"nationalTeamActive"`
	NationalTeamID     int             `json:"-"`
	LastTradePrice     map[int]float64 `json:"lastTradePrice"`
	LastBotCycleAt     string          `json:"lastBotCycleAt"`
}

// GetCompany returns a pointer to the company with the given ID.
func (gs *GameState) GetCompany(id int) *Company {
	for i := range gs.Companies {
		if gs.Companies[i].ID == id {
			return &gs.Companies[i]
		}
	}
	return nil
}
