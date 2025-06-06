package notification

type PushNotificationRequest struct {
	UserId   int64             `json:"user_id,omitempty"`
	DeviceId string            `json:"device_id,omitempty"`
	Title    string            `json:"title,omitempty"`
	Body     string            `json:"body,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

type EmailNotificationRequest struct {
	UserId  int64             `json:"user_id,omitempty"`
	Email   string            `json:"email,omitempty"`
	Subject string            `json:"subject,omitempty"`
	Body    string            `json:"body,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
}
