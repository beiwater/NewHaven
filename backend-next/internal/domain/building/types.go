package building

// CatalogEntry is a building type from the static catalog (buildings.json).
type CatalogEntry struct {
	ID          int    `json:"id"`
	Kind        int    `json:"kind"`
	Name        string `json:"name"`
	Type        string `json:"type"` // "production" or "retail"
	BaseCost    int    `json:"base_cost"`
	BaseOutput  int    `json:"base_output"`
	Produces    []int  `json:"produces,omitempty"`
	Description string `json:"description,omitempty"`
}

// ResourceCatalogEntry is a resource type from resources.json.
type ResourceCatalogEntry struct {
	ID                  int     `json:"id"`
	DbLetter            int     `json:"db_letter"`
	Name                string  `json:"name"`
	Category            string  `json:"category"` // "raw", "processed", "terminal"
	Tier                int     `json:"tier"`
	ProducedPerHourRaw  int     `json:"produced_per_hour_raw"`
	UnitsSoldAnHour     int     `json:"units_sold_an_hour"`
	IsExchangeTradable  bool    `json:"is_exchange_tradable"`
	HasEconomyModel     bool    `json:"has_economy_model"`
	BasePrice           float64 `json:"base_price"`
}
