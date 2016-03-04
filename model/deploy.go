package model

type Deploy struct {
	Type       string `json:"type"`
	Deployment *Deployment
}
