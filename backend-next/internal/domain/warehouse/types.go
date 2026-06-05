package warehouse

// Warehouse represents a company's warehouse storage.
type Warehouse struct {
	CompanyID    int    `json:"company_id"`
	Capacity     int    `json:"capacity"`
	UsedCapacity int    `json:"used_capacity"`
	Items        []Item `json:"items"`
}

// Item represents a single stack of resources in the warehouse.
type Item struct {
	ResourceID   int    `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	Quality      int    `json:"quality"`
	Amount       int    `json:"amount"`
}
