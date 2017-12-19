package repository

import "fmt"

// QueueNotFoundError is an error returned when a non-existing queue
// is specifie as the destination of a routing.
type QueueNotFoundError struct {
	QueueName string
}

func (qe *QueueNotFoundError) Error() string {
	return fmt.Sprintf("No such queue: %s", qe.QueueName)
}
