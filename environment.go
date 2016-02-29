package core

type Environment struct {
	Client *Client
	Name   string
}

// Unmarshal is required by the Entity interface
func (e *Environment) Unmarshal(n *Node) error {
	return n.ValueInto(e)
}
