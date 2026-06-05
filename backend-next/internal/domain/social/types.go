package social

// Message represents a chat message.
type Message struct {
	ID         int64  `json:"id"`
	CompanyID  int    `json:"company_id"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	Channel    string `json:"channel"` // "global", "company", "system"
	CreatedAt  string `json:"created_at"`
}

// Notification is a system notification for a company.
type Notification struct {
	ID        int    `json:"id"`
	CompanyID int    `json:"company_id"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}
