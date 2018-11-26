package checker

// Pair is a key-value pair to hold intermediate state collection results.
type Pair struct {
	Key   string
	Value string
}

func (p Pair) String() string {
	return p.Key + ":" + p.Value
}

// Checker types which implement different types of checks.
type Checker interface {
	// Gather info about the relevant state of the system. (async)
	Collect(config map[string]string)
	// Get current progress of the collection.
	Progress() string
	// Fetch the collected system state info.
	GetCollected() ([]Pair, error)
	// Get error if any has occured during collection.
	GetErr() error
}

type BasicChecker struct {
	// Tracks progress of the collection operation, since some can take a while.
	progress string
	// Collected state of the system
	collected []Pair
	err       error
}
