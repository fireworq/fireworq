package jobqueue

// Node describes information of an active queue node.
type Node struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

// HasNodeInfo is an interface describing that it has a Node information.
//
// This is typically a JobQueue sub-interface.
type HasNodeInfo interface {
	Node() (*Node, error)
}
