package worker

import (
	"errors"

	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/pipeline"
)

type work struct {
	Description        pipeline.Pipeline
	File               string
	Bucket             string
	LocalDir           string
	ResultsDestination string
}

func (w work) getLockKey() string {
	return w.Bucket + ":" + w.File
}

type workQueue struct {
	memory []work
	curPos int
}

func makeWorkQueue() *workQueue {
	return &workQueue{
		curPos: -1,
		memory: make([]work, 0),
	}
}

func (queue *workQueue) Add(e work) {
	queue.memory = append(queue.memory, e)
}

func (queue *workQueue) Pop() (work, error) {
	if len(queue.memory) == 0 {
		return work{}, errors.New("the queue is empty")
	}
	e := queue.memory[0]
	queue.memory = queue.memory[1:]
	return e, nil
}

func (queue *workQueue) Clear() {
	queue.memory = make([]work, 0)
}

func (queue *workQueue) getNext() (work, error) {
	if queue.curPos == len(queue.memory)-1 {
		return work{}, errors.New("end of queue")
	}
	queue.curPos++
	return queue.memory[queue.curPos], nil
}

func (queue *workQueue) choseCurrent() {
	if queue.curPos == len(queue.memory)-1 {
		queue.memory = queue.memory[0:queue.curPos]
	} else if queue.curPos > 0 {
		queue.memory = append(queue.memory[0:queue.curPos], queue.memory[queue.curPos+1:]...)
	} else if queue.curPos == 0 {
		queue.memory = queue.memory[1:]
	}
}
