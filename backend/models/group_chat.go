package models

// GroupChat represents a group chat
type GroupChat struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedBy int64  `json:"created_by"` // User ID who created the group chat
	CreatedAt string `json:"created_at"`
}
