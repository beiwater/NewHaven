package memory

import (
	"maps"
	"slices"

	"github.com/beiwater/NewHaven/backend/internal/domain/company"
)

func cloneCompany(source *company.Company) *company.Company {
	cloned := *source
	cloned.Preferences = cloneJSONMap(source.Preferences)
	cloned.Buildings = slices.Clone(source.Buildings)
	for i := range cloned.Buildings {
		cloned.Buildings[i].Shelves = slices.Clone(source.Buildings[i].Shelves)
	}
	cloned.Inventory = maps.Clone(source.Inventory)
	cloned.Executives = slices.Clone(source.Executives)
	cloned.RetailCarry = maps.Clone(source.RetailCarry)
	return &cloned
}

func cloneJSONMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = cloneJSONValue(value)
	}
	return cloned
}

func cloneJSONValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		return cloneJSONMap(value)
	case []any:
		cloned := make([]any, len(value))
		for i, item := range value {
			cloned[i] = cloneJSONValue(item)
		}
		return cloned
	default:
		return value
	}
}
