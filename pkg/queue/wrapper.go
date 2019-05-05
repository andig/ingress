package queue

import (
	"errors"
	"fmt"

	equeue "github.com/eapache/queue"
)

// Queue is a fast, ring-buffered queue
type Queue struct {
	*equeue.Queue
}

// New creates Queue
func New() *Queue {
	return &Queue{equeue.New()}
}

func recoverWithError() error {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			return errors.New(x)
		case error:
			return x
		default:
			return errors.New("unknown panic")
		}
	}
	return nil
}

// Peek into queue top
func (q *Queue) Peek() (res interface{}, err error) {
	defer func() {
		err = recoverWithError()
	}()

	return q.Queue.Peek(), nil
}

// Get queue element by index
func (q *Queue) Get(i int) (res interface{}, err error) {
	defer func() {
		err = recoverWithError()
	}()

	return q.Queue.Get(i), nil
}

// Remove top element from queue
func (q *Queue) Remove() (res interface{}, err error) {
	defer func() {
		err = recoverWithError()
	}()

	return q.Queue.Remove(), nil
}

// String representation of queue
func (q *Queue) String() string {
	s := "["
	for i := 0; i < q.Length(); i++ {
		if i > 0 {
			s += ", "
		}
		v, _ := q.Get(i)
		s += fmt.Sprintf("%v", v)
	}
	return s + "]"
}
