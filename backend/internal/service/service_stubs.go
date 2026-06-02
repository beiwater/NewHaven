package service
import (
	"fmt"
	"math"
	"sync"
	"time"
)

// ── Executive model ───────────────────────────────────

type ExecutiveStub struct {
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
	RecruitCost     float64 `json:"recruitCost"`
	TrainingCost    float64 `json:"trainingCost"`
	TrainingTime    float64 `json:"trainingTime"` // seconds
	Status          string  `json:"status"`        // "idle" | "training" | "recruiting"
	TrainingEndTime *string `json:"trainingEndTime,omitempty"`
}

// ── Curve helpers ──

func productionBonusAtLevel(level int) float64 {
	if level < 1 {
		return 0
	}
	base := 2.0
	increment := 0.9 - float64(level)*0.06
	raw := base + float64(level-1)*math.Max(0.12, increment)
	return math.Round(raw*10) / 10
}

func salesBonusAtLevel(level int) float64 {
	if level < 1 {
		return 0
	}
	base := 4.0
	increment := 1.6 - float64(level)*0.08
	raw := base + float64(level-1)*math.Max(0.2, increment)
	return math.Round(raw*10) / 10
}

func mgmtDiscountAtLevel(level int) float64 {
	if level < 1 {
		return 0
	}
	base := 2.0
	increment := 0.7 - float64(level)*0.035
	raw := base + float64(level-1)*math.Max(0.1, increment)
	return math.Round(raw*10) / 10
}

func stageAtLevel(level int) string {
	switch {
	case level >= 10:
		return "Executive VP"
	case level >= 8:
		return "Director"

	case level >= 6:
		return "Senior Manager"
	case level >= 4:
		return "Manager"
	case level >= 2:
		return "Junior Manager"
	default:
		return "Trainee"
	}
}

func trainingCost(level int) float64 {
	return math.Round(5000 * math.Pow(float64(level), 1.6))
}

func trainingTimeSeconds(level int) float64 {
	return math.Round(3600 * math.Pow(float64(level), 0.7))
}

func salaryAtLevel(level int) float64 {
	return math.Round(600 + 80*math.Pow(float64(level), 1.3))
}

func recruitCost(rarity string, level int) float64 {
	rarityFactor := map[string]float64{
		"Legendary": 2.5,
		"Epic":      1.8,
		"Rare":      1.2,
		"Common":    0.8,
	}
	factor := rarityFactor[rarity]
	if factor == 0 {
		factor = 1.0
	}
	return math.Round(15000 * factor * math.Pow(float64(level), 0.8))
}

func inferRarity(level int) string {
	switch {
	case level >= 9:
		return "Legendary"
	case level >= 7:
		return "Epic"
	case level >= 5:
		return "Rare"
	default:
		return "Common"
	}
}

// ── Market generation ─────────────────────────────────

var executiveNames = []struct {
	name  string
	title string
	level int
}{
	{"Sophia Grant", "Sales Director", 8},
	{"Marcus Vale", "Operations Manager", 6},
	{"Elena Brooks", "VP Operations", 10},
	{"Alice Chen", "Chief Operating Officer", 7},
	{"Bob Smith", "Chief Technology Officer", 5},
	{"Carol Davis", "Chief Financial Officer", 9},
	{"James Miller", "Chief Financial Officer", 7},
	{"Olivia Bennett", "Supply Chain Director", 5},
	{"Daniel Park", "Marketing Director", 6},
	{"Julia Torres", "Production Manager", 4},
}

func generateMarketExecutives() []map[string]any {
	execs := make([]map[string]any, 0, 3)
	// pick 3 random-like entries
	for i := 0; i < 3 && i < len(executiveNames); i++ {
		e := executiveNames[i]
		lv := e.level
		rarity := inferRarity(lv)

		execs = append(execs, map[string]any{
			"id":              "exec-" + string(rune('a'+i)),
			"name":            e.name,
			"title":           e.title,
			"level":           lv,
			"rarity":          rarity,
			"stage":           stageAtLevel(lv),
			"salary":          salaryAtLevel(lv),
			"productionBonus": productionBonusAtLevel(lv),
			"salesBonus":      salesBonusAtLevel(lv),
			"mgmtDiscount":    mgmtDiscountAtLevel(lv),
			"recruitCost":     recruitCost(rarity, lv),
			"trainingCost":    trainingCost(lv),
			"trainingTime":    trainingTimeSeconds(lv),
			"status":          "idle",
		})
	}
	return execs
}

// ── Per-instance state (owned by Service) ──────────────

type execState struct {
	mu          sync.Mutex
	hired       map[string]*ExecutiveStub
	market      []*ExecutiveStub
	lastRefresh time.Time
}

// newExecState creates a fresh execState for a Service instance.
func newExecState() *execState {
	return &execState{hired: make(map[string]*ExecutiveStub)}
}

func (s *Service) MyExecutives() []map[string]any {
	if s.execState == nil {
		return nil
	}
	s.execState.mu.Lock()
	defer s.execState.mu.Unlock()

	result := make([]map[string]any, 0, len(s.execState.hired))
	for _, e := range s.execState.hired {
		result = append(result, stubToMap(e))
	}
	return result
}

// ── Service methods (called by handlers) ──────────────

func (s *Service) ExecutiveCatalog() []map[string]any {
	if s.execState == nil {
		return nil
	}
	s.execState.mu.Lock()
	defer s.execState.mu.Unlock()

	// Refresh market periodically
	if time.Since(s.execState.lastRefresh) > 10*time.Minute || len(s.execState.market) == 0 {
		raw := generateMarketExecutives()
		s.execState.market = make([]*ExecutiveStub, len(raw))
		for i, r := range raw {
			s.execState.market[i] = stubFromMap(r)
		}
		s.execState.lastRefresh = time.Now()
	}

	result := make([]map[string]any, len(s.execState.market))
	for i, e := range s.execState.market {
		result[i] = stubToMap(e)
	}
	return result
}

func (s *Service) RecruitExecutive(execID string) map[string]any {
	if s.execState == nil {
		return map[string]any{"ok": false, "error": "exec state not initialized"}
	}
	s.execState.mu.Lock()
	defer s.execState.mu.Unlock()

	// Find in market
	for _, e := range s.execState.market {
		if e.ID == execID {
			// Check if already hired
			if _, exists := s.execState.hired[execID]; exists {
				return map[string]any{"ok": false, "error": "already recruited"}
			}
			// Clone to hired
			clone := *e
			clone.Status = "idle"
			clone.TrainingEndTime = nil
			s.execState.hired[execID] = &clone
			return map[string]any{"ok": true, "executive": stubToMap(&clone)}
		}
	}
	return map[string]any{"ok": false, "error": "executive not found"}
}

func (s *Service) TrainExecutive(execID string) map[string]any {
	if s.execState == nil {
		return map[string]any{"ok": false, "error": "exec state not initialized"}
	}
	s.execState.mu.Lock()
	defer s.execState.mu.Unlock()

	e, exists := s.execState.hired[execID]
	if !exists {
		return map[string]any{"ok": false, "error": "executive not found"}
	}

	if e.Status == "training" {
		// Check if training is done
		if e.TrainingEndTime != nil {
			end, err := time.Parse(time.RFC3339, *e.TrainingEndTime)
			if err == nil && time.Now().After(end) {
				// Complete training
				e.Level++
				e.Status = "idle"
				e.TrainingEndTime = nil
				updateExecutiveStats(e)
				return map[string]any{"ok": true, "executive": stubToMap(e), "trained": true}
			}
		}
		return map[string]any{"ok": false, "error": "executive is already training"}
	}

	// Start training
	cost := trainingCost(e.Level)
	e.Status = "training"
	duration := trainingTimeSeconds(e.Level)
	end := time.Now().Add(time.Duration(duration) * time.Second)
	endStr := end.Format(time.RFC3339)
	e.TrainingEndTime = &endStr

	return map[string]any{"ok": true, "executive": stubToMap(e), "cost": cost, "trainingTime": duration}
}

func updateExecutiveStats(e *ExecutiveStub) {
	e.Stage = stageAtLevel(e.Level)
	e.Rarity = inferRarity(e.Level)
	e.Salary = salaryAtLevel(e.Level)
	e.ProductionBonus = productionBonusAtLevel(e.Level)
	e.SalesBonus = salesBonusAtLevel(e.Level)
	e.MgmtDiscount = mgmtDiscountAtLevel(e.Level)
	e.TrainingCost = trainingCost(e.Level)
	e.TrainingTime = trainingTimeSeconds(e.Level)
}

func (s *Service) ExecutiveDetail(execID string) (map[string]any, error) {
	if s.execState == nil {
		return nil, fmt.Errorf("executive state not initialized")
	}
	s.execState.mu.Lock()
	defer s.execState.mu.Unlock()

	// Check hired first
	if e, exists := s.execState.hired[execID]; exists {
		return stubToMap(e), nil
	}
	// Check market
	for _, e := range s.execState.market {
		if e.ID == execID {
			m := stubToMap(e)
			m["morale"] = 100
			return m, nil
		}
	}
	return nil, fmt.Errorf("executive not found: %s", execID)
}

func (s *Service) PoachExecutive(_ map[string]any) map[string]any {
	return map[string]any{"ok": false, "reason": "poach_failed"}
}

func (s *Service) IncomingOffers() []map[string]any {
	return []map[string]any{}
}

func (s *Service) RespondToOffer(_ map[string]any) map[string]any {
	return map[string]any{"ok": true}
}

func stubFromMap(m map[string]any) *ExecutiveStub {
	return &ExecutiveStub{
		ID:              getStr(m, "id"),
		Name:            getStr(m, "name"),
		Title:           getStr(m, "title"),
		Level:           int(getFloat(m, "level")),
		Rarity:          getStr(m, "rarity"),
		Stage:           getStr(m, "stage"),
		Salary:          getFloat(m, "salary"),
		ProductionBonus: getFloat(m, "productionBonus"),
		SalesBonus:      getFloat(m, "salesBonus"),
		MgmtDiscount:    getFloat(m, "mgmtDiscount"),
		RecruitCost:     getFloat(m, "recruitCost"),
		TrainingCost:    getFloat(m, "trainingCost"),
		TrainingTime:    getFloat(m, "trainingTime"),
		Status:          getStr(m, "status"),
	}
}

func stubToMap(e *ExecutiveStub) map[string]any {
	m := map[string]any{
		"id":              e.ID,
		"name":            e.Name,
		"title":           e.Title,
		"level":           e.Level,
		"rarity":          e.Rarity,
		"stage":           e.Stage,
		"salary":          e.Salary,
		"productionBonus": e.ProductionBonus,
		"salesBonus":      e.SalesBonus,
		"mgmtDiscount":    e.MgmtDiscount,
		"recruitCost":     e.RecruitCost,
		"trainingCost":    e.TrainingCost,
		"trainingTime":    e.TrainingTime,
		"status":          e.Status,
	}
	if e.TrainingEndTime != nil {
		m["trainingEndTime"] = *e.TrainingEndTime
	}
	return m
}


func getStr(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func getFloat(m map[string]any, key string) float64 {
	v, _ := m[key].(float64)
	return v
}

// ── Stubs for Aerospace (unchanged) ───────────────────

func (s *Service) RocketProjects() []map[string]any {
	return []map[string]any{
		{"id": "rocket-1", "name": "Atlas I", "type": "cargo", "progress": 45, "cost": 500000, "timeLeft": "72h"},
		{"id": "rocket-2", "name": "Pioneer X", "type": "passenger", "progress": 12, "cost": 1200000, "timeLeft": "240h"},
	}
}

func (s *Service) CreateRocketProject(_ map[string]any) map[string]any {
	return map[string]any{"ok": true, "id": "rocket-new", "name": "New Rocket"}
}

func (s *Service) LaunchHistory() []map[string]any {
	return []map[string]any{
		{"id": "launch-1", "rocket": "Atlas I", "destination": "Orbit", "status": "success", "profit": 25000},
	}
}

func (s *Service) LaunchRocket(_ map[string]any) map[string]any {
	return map[string]any{"ok": true, "launchId": "launch-new"}
}

func (s *Service) AvailableComponents() []map[string]any {
	return []map[string]any{
		{"id": "comp-1", "name": "Solar Panel MK3", "type": "energy", "cost": 5000},
		{"id": "comp-2", "name": "Thruster V2", "type": "propulsion", "cost": 12000},
	}
}
