package queue

import (
	"errors"

	pqueue "github.com/eapache/queue"
)

type Queue struct {
	*pqueue.Queue
}

func New() *Queue {
	return &Queue{pqueue.New()}
}

func recoverWithError() error {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			return errors.New(x)
		case error:
			return x
		default:
			return errors.New("Unknown panic")
		}
	}
	return nil
}

func (q *Queue) Peek() (res interface{}, err error) {
	defer func() {
		err = recoverWithError()
	}()

	return q.Queue.Peek(), nil
}

func (q *Queue) Get(i int) (res interface{}, err error) {
	defer func() {
		err = recoverWithError()
	}()

	return q.Queue.Get(i), nil
}

func (q *Queue) Remove() (res interface{}, err error) {
	defer func() {
		err = recoverWithError()
	}()

	return q.Queue.Remove(), nil
}
