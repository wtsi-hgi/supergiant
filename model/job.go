package model

type Job struct {
	ID       string `json:"-"`
	Type     int    `json:"type"`
	Data     string `json:"data"`
	Status   string `json:"status"`
	Attempts int    `json:"attempts"`
	Error    string `json:"error"`
}
