package points

// Action type represents store for action cost
type Action struct {
	cost int64
}

// Creates new Action instance
func NewAction() Action {
	return Action{}
}

// Add cost to action
func (a *Action) Add(points int64) {
	a.cost += points
}

// Get the action cost
func (a Action) Cost() int64 {
	return a.cost
}
