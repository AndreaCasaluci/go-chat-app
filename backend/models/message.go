package models

import "github.com/google/uuid"

// Message represents a message in the system
type Message struct {
	ID          int64     `json:"id"`
	UUID        uuid.UUID `json:"uuid"`
	SenderID    int64     `json:"sender_id"`
	ReceiverID  int64     `json:"receiver_id"`
	GroupID     int64     `json:"group_id"`
	MessageText string    `json:"message_text"`
	MediaType   string    `json:"media_type"` // text, image, video
	MediaURL    string    `json:"media_url"`  // URL for media
	CreatedAt   string    `json:"created_at"`
}
