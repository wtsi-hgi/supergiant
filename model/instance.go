package model

// Instance is really just a Kubernetes Pod (with a better name)
type Instance struct {
	ID     int  `json:"id"` // actually just the number (starting w/ 1) of the instance order in the deployment
	Active bool `json:"active"`
}
