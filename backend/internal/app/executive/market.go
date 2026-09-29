package executive

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/domain/company"
)

const candidatesPerRefresh = 8

// Candidate is a deterministic hourly market offer. Its ID contains the
// market hour and slot so recruitment can always validate the exact candidate
// the player clicked instead of rebuilding the first random executive.
type Candidate struct {
	company.Executive
	RecruitCost int `json:"recruitCost"`
}

var candidateNames = []string{
	"Alice Chen", "Bob Martinez", "Carol Wu", "David Park",
	"Elena Rossi", "Frank Okafor", "Grace Kim", "Henry Dubois",
	"Iris Johansson", "James Tanaka", "Kate Schmidt", "Leo Andersen",
}

var executiveRoles = []company.ExecutivePosition{
	company.ExecutivePositionCOO,
	company.ExecutivePositionCFO,
	company.ExecutivePositionCMO,
	company.ExecutivePositionCTO,
}

var rarityWeights = []struct {
	name   string
	weight int
}{
	{"Common", 50},
	{"Rare", 30},
	{"Epic", 15},
	{"Legendary", 5},
}

func candidatesForHour(now time.Time) []Candidate {
	hour := now.UTC().Unix() / 3600
	rng := rand.New(rand.NewSource(hour*7919 + 17))
	candidates := make([]Candidate, 0, candidatesPerRefresh)
	for slot := 0; slot < candidatesPerRefresh; slot++ {
		candidates = append(candidates, candidateForSlot(rng, hour, slot))
	}
	return candidates
}

func candidateByID(now time.Time, id string) (Candidate, bool) {
	parts := strings.Split(id, "-")
	if len(parts) != 4 || parts[0] != "exec" || parts[1] != "market" {
		return Candidate{}, false
	}
	hour, hourErr := strconv.ParseInt(parts[2], 10, 64)
	slot, slotErr := strconv.Atoi(parts[3])
	if hourErr != nil || slotErr != nil || slot < 0 || slot >= candidatesPerRefresh {
		return Candidate{}, false
	}
	currentHour := now.UTC().Unix() / 3600
	if hour != currentHour {
		return Candidate{}, false
	}
	for _, candidate := range candidatesForHour(now) {
		if candidate.ID == id {
			return candidate, true
		}
	}
	return Candidate{}, false
}

func candidateForSlot(rng *rand.Rand, hour int64, slot int) Candidate {
	level := rng.Intn(8) + 1
	rarity := pickRarity(rng.Intn(100))
	specialty := executiveRoles[rng.Intn(len(executiveRoles))]
	skills := generatedSkills(rng, level, specialty, rarity)
	executive := company.Executive{
		ID:        fmt.Sprintf("exec-market-%d-%d", hour, slot),
		Name:      candidateNames[rng.Intn(len(candidateNames))],
		Title:     executiveTitle(specialty),
		Specialty: specialty,
		Skills:    skills,
		Level:     level,
		Rarity:    rarity,
		Stage:     stageAtLevel(level),
		Salary:    salaryAtLevel(level),
		Morale:    100,
	}
	executive.ProductionBonus = productionBonusFor(&executive)
	executive.SalesBonus = salesBonusFor(&executive)
	executive.MgmtDiscount = managementBonusFor(&executive)
	return Candidate{Executive: executive, RecruitCost: RecruitCost(rarity, level, skills)}
}

func generatedSkills(rng *rand.Rand, level int, specialty company.ExecutivePosition, rarity string) company.ExecutiveSkills {
	base := float64(10 + level*4 + rng.Intn(8))
	rarityBonus := map[string]float64{"Common": 0, "Rare": 5, "Epic": 11, "Legendary": 18}[rarity]
	skills := company.ExecutiveSkills{
		Management:    clampSkill(base + rarityBonus + float64(rng.Intn(9)-4)),
		Accounting:    clampSkill(base + rarityBonus + float64(rng.Intn(9)-4)),
		Communication: clampSkill(base + rarityBonus + float64(rng.Intn(9)-4)),
		Science:       clampSkill(base + rarityBonus + float64(rng.Intn(9)-4)),
	}
	primaryBoost := 12 + rarityBonus/2
	switch specialty {
	case company.ExecutivePositionCOO:
		skills.Management = clampSkill(skills.Management + primaryBoost)
	case company.ExecutivePositionCFO:
		skills.Accounting = clampSkill(skills.Accounting + primaryBoost)
	case company.ExecutivePositionCMO:
		skills.Communication = clampSkill(skills.Communication + primaryBoost)
	case company.ExecutivePositionCTO:
		skills.Science = clampSkill(skills.Science + primaryBoost)
	}
	return skills
}

func clampSkill(value float64) float64 {
	return math.Max(1, math.Min(100, math.Round(value)))
}

func RecruitCost(rarity string, level int, skills company.ExecutiveSkills) int {
	if level < 1 {
		level = 1
	}
	factor := map[string]float64{"Common": 0.8, "Rare": 1.2, "Epic": 1.8, "Legendary": 2.5}[rarity]
	if factor == 0 {
		factor = 1
	}
	averageSkill := (skills.Management + skills.Accounting + skills.Communication + skills.Science) / 4
	return int(math.Round(15000 * factor * math.Pow(float64(level), 0.8) * (0.75 + averageSkill/200)))
}

func pickRarity(roll int) string {
	total := 0
	for _, rarity := range rarityWeights {
		total += rarity.weight
	}
	roll %= total
	cumulative := 0
	for _, rarity := range rarityWeights {
		cumulative += rarity.weight
		if roll < cumulative {
			return rarity.name
		}
	}
	return "Common"
}
