package market

// OrderSide represents buy or sell.
type OrderSide string

const (
	Buy  OrderSide = "buy"
	Sell OrderSide = "sell"
)

// OrderStatus represents order lifecycle.
type OrderStatus string

const (
	StatusOpen      OrderStatus = "open"
	StatusPartial   OrderStatus = "partial"
	StatusFilled    OrderStatus = "filled"
	StatusCancelled OrderStatus = "cancelled"
)

// MarketOrder is a limit or market order on the exchange.
type MarketOrder struct {
	ID             string      `json:"id"`
	CompanyID      int         `json:"company_id"`
	ResourceID     int         `json:"resource_id"`
	IsBuy          bool        `json:"is_buy"`
	Price          float64     `json:"price"`
	Quantity       int         `json:"quantity"`
	FilledQuantity int         `json:"filled_quantity"`
	Quality        int         `json:"quality"`
	Status         OrderStatus `json:"status"`
	CreatedAt      string      `json:"created_at"`
}

// Remaining returns the unfilled quantity.
func (o *MarketOrder) Remaining() int {
	r := o.Quantity - o.FilledQuantity
	if r < 0 {
		return 0
	}
	return r
}

// Trade represents an executed trade between two orders.
type Trade struct {
	ID          string  `json:"id"`
	BuyOrderID  string  `json:"buy_order_id"`
	SellOrderID string  `json:"sell_order_id"`
	ResourceID  int     `json:"resource_id"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	BuyerFee    float64 `json:"buyer_fee"`
	SellerFee   float64 `json:"seller_fee"`
	CreatedAt   string  `json:"created_at"`
}

// Ticker represents the current market state for a resource.
type Ticker struct {
	ResourceID  int     `json:"resource_id"`
	LastPrice   float64 `json:"last_price"`
	Volume24h   float64 `json:"volume_24h"`
	High24h     float64 `json:"high_24h"`
	Low24h      float64 `json:"low_24h"`
	PriceChange float64 `json:"price_change_24h"`
	UpdatedAt   string  `json:"updated_at"`
}

// DepthLevel represents aggregated supply/demand at a price level.
type DepthLevel struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Orders   int     `json:"orders"`
}

// OrderbookDepth shows the current order book.
type OrderbookDepth struct {
	ResourceID int          `json:"resource_id"`
	Bids       []DepthLevel `json:"bids"`
	Asks       []DepthLevel `json:"asks"`
}

// CreateOrderRequest is the DTO for creating a new order.
type CreateOrderRequest struct {
	ResourceID int     `json:"resource_id" validate:"required,min=1"`
	IsBuy      bool    `json:"is_buy"`
	Price      float64 `json:"price" validate:"required,gt=0"`
	Quantity   int     `json:"quantity" validate:"required,min=1"`
}
