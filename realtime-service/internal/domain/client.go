package domain

type Client struct {
	GroupID string
	UserID  string
	Send    chan Event
}

func NewClient(groupID, userID string) *Client {
	return &Client{
		GroupID: groupID,
		UserID:  userID,
		Send:    make(chan Event, 32),
	}
}
