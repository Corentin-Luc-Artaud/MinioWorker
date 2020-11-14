package network

import "errors"

type requestType string

const (
	download = requestType("download")
	upload   = requestType("upload")
)

type networkRequest struct {
	requestType requestType
	localpath   string
	remotepath  string
	bucket      string
	notifChan   chan error
}

//requestQueue non thread safe fifo
type requestQueue struct {
	memory []networkRequest
}

func (queue *requestQueue) Add(e networkRequest) {
	queue.memory = append(queue.memory, e)
}

func (queue *requestQueue) Pop() (networkRequest, error) {
	if len(queue.memory) == 0 {
		return networkRequest{}, errors.New("the queue is empty")
	}
	e := queue.memory[0]
	queue.memory = queue.memory[1:]
	return e, nil
}

func (queue *requestQueue) Clear() {
	queue.memory = []networkRequest{}
}
