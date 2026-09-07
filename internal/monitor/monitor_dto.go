package monitor

import "fmt"

type MonitorDTO struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	URL                   string  `json:"url"`
	Method                string  `json:"method"`
	IntervalSeconds       int     `json:"interval_seconds"`
	TimeoutSeconds        int     `json:"timeout_seconds"`
	ExpectedStatus        int     `json:"expected_status"`
	ExpectedBodySubstring *string `json:"expected_body_substring,omitempty"`
	HeadersJSON           *string `json:"headers_json,omitempty"`
	Enabled               bool    `json:"enabled"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

func (m MonitorDTO) Debug() {
	fmt.Printf("%+v\n", m)
}
