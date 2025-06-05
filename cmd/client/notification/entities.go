package notification

type PushNotificationRequest struct {
	UserId int64             `json:"user_id,omitempty"`
	Title  string            `json:"title,omitempty"`
	Body   string            `json:"body,omitempty"`
	Data   map[string]string `json:"data,omitempty"`
}
