package bridge

import (
	"encoding/json"
	"fmt"
)

// ComparisonResult describes the diff between old and new backend responses.
type ComparisonResult struct {
	Route       string `json:"route"`
	OldStatus   int    `json:"old_status"`
	NewStatus   int    `json:"new_status"`
	MatchLevel  string `json:"match_level"` // "status", "shape", "semantic", "exact"
	Match       bool   `json:"match"`
	DiffSummary string `json:"diff_summary,omitempty"`
}

// CompareResponses compares two JSON responses at a given level of strictness.
func CompareResponses(route string, oldStatus int, oldBody []byte, newStatus int, newBody []byte, level string) ComparisonResult {
	result := ComparisonResult{
		Route:      route,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		MatchLevel: level,
	}

	if oldStatus != newStatus {
		result.Match = false
		result.DiffSummary = fmt.Sprintf("status mismatch: %d vs %d", oldStatus, newStatus)
		return result
	}

	// Status-only comparison
	if level == "status" {
		result.Match = true
		return result
	}

	// JSON comparison
	var oldJSON, newJSON any
	if err := json.Unmarshal(oldBody, &oldJSON); err != nil {
		result.Match = false
		result.DiffSummary = fmt.Sprintf("old response not JSON: %v", err)
		return result
	}
	if err := json.Unmarshal(newBody, &newJSON); err != nil {
		result.Match = false
		result.DiffSummary = fmt.Sprintf("new response not JSON: %v", err)
		return result
	}

	if level == "exact" {
		oldStr, _ := json.Marshal(oldJSON)
		newStr, _ := json.Marshal(newJSON)
		if string(oldStr) != string(newStr) {
			result.Match = false
			result.DiffSummary = "responses differ"
			return result
		}
		result.Match = true
		return result
	}

	if level == "shape" {
		if !keysMatch(oldJSON, newJSON) {
			result.Match = false
			result.DiffSummary = "key sets differ"
			return result
		}
	}

	// shape and semantic pass if we got here
	result.Match = true
	return result
}

func keysMatch(a, b any) bool {
	am, okA := a.(map[string]any)
	bm, okB := b.(map[string]any)
	if okA && okB {
		for k := range am {
			if _, exists := bm[k]; !exists {
				return false
			}
			if !keysMatch(am[k], bm[k]) {
				return false
			}
		}
		return true
	}
	return true // for non-object types, skip recursive key check
}
