package chat

type ChatRoom struct {
	ID            string `json:"id"`           // "p:1000001-1000002"
	Participant1  int    `json:"participant1"` // smaller company ID
	Participant2  int    `json:"participant2"` // larger company ID
	LastMessageAt string `json:"last_message_at,omitempty"`
	LastMessage   string `json:"last_message,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type Message struct {
	ID         int64  `json:"id"`
	RoomID     string `json:"room_id"`
	SenderID   int    `json:"sender_id"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}
