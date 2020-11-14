package util

import "errors"

//Queue non thread safe fifo
type Queue struct {
	memory []string
}

func (queue *Queue) Add(e string) {
	queue.memory = append(queue.memory, e)
}

func (queue *Queue) Pop() (string, error) {
	if len(queue.memory) == 0 {
		return "", errors.New("the queue is empty")
	}
	e := queue.memory[0]
	queue.memory = queue.memory[1:]
	return e, nil
}

func (queue *Queue) Clear() {
	queue.memory = []string{}
}
