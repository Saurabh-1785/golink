package proto

// RecordClickRequest is sent when a short URL is accessed.
type RecordClickRequest struct {
	ShortCode string `json:"short_code"`
}

// RecordClickResponse confirms the click was recorded.
type RecordClickResponse struct {
	Success bool `json:"success"`
}

// GetStatsRequest asks for analytics of a short code.
type GetStatsRequest struct {
	ShortCode string `json:"short_code"`
}

// GetStatsResponse returns the click count and last clicked time.
type GetStatsResponse struct {
	ShortCode     string `json:"short_code"`
	ClickCount    int64  `json:"click_count"`
	LastClickedAt string `json:"last_clicked_at,omitempty"`
}
