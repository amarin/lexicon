package nertest

// Gold is one expected span.
type Gold struct {
	Text       string `json:"text"`
	Type       string `json:"type"`
	Occurrence int    `json:"occurrence,omitempty"` // 1-based, default 1
}
