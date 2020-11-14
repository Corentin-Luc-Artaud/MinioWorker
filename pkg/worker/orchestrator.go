package worker

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/network"
	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/pipeline"
	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/util"

	"github.com/minio/minio-go/v7/pkg/notification"
)

//MinIOWorker is the object responsible of orchestrating workload
type MinIOWorker struct {
	maxConcurentWorker     int
	currentConcurentWorker int
	workingTarget          *util.Set
	workQueue              *workQueue
	endWorker              chan string
	eventChan              <-chan notification.Info
	handlePipelines        []pipeline.Pipeline
	networkManager         *network.NetworkManager
	ResultsDestination     string
}

//NewMinIOWorker instantiate a new orchestrator
func NewMinIOWorker(eventChan <-chan notification.Info, maxConcurentWorker int, handlePipelines []pipeline.Pipeline, networkManager *network.NetworkManager, ResultsDestination string) *MinIOWorker {
	works := makeWorkQueue()
	return &MinIOWorker{
		currentConcurentWorker: 0,
		maxConcurentWorker:     maxConcurentWorker,
		workingTarget:          util.NewSet(),
		workQueue:              works,
		endWorker:              make(chan string),
		eventChan:              eventChan,
		handlePipelines:        handlePipelines,
		networkManager:         networkManager,
		ResultsDestination:     ResultsDestination,
	}
}

//Run main loop of the Orchestrator
func (minIOWorker *MinIOWorker) Run() error {
	for {
		select {
		case event := <-minIOWorker.eventChan:
			if event.Err != nil {
				return event.Err
			}
			if err := minIOWorker.handleRecords(event.Records); err != nil {
				return err
			}
		case end := <-minIOWorker.endWorker:
			minIOWorker.handleEnd(end)
		}
	}
}

func (minIOWorker *MinIOWorker) handleRecords(records []notification.Event) error {
	for _, record := range records {
		bucket := record.S3.Bucket.Name
		file := record.S3.Object.Key
		err := minIOWorker.triggerOrEnqueueWork(bucket, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (minIOWorker *MinIOWorker) triggerOrEnqueueWork(bucket string, file string) error {
	for _, handlePipeline := range minIOWorker.handlePipelines {
		if !handlePipeline.IsHandelable(bucket, file) {
			continue
		}

		w := work{
			Bucket:             bucket,
			File:               file,
			Description:        handlePipeline,
			LocalDir:           getTempdirectory(),
			ResultsDestination: minIOWorker.ResultsDestination,
		}
		//trigger if possible
		if minIOWorker.currentConcurentWorker < minIOWorker.maxConcurentWorker && !minIOWorker.workingTarget.Contains(w.getLockKey()) {
			minIOWorker.currentConcurentWorker++
			minIOWorker.workingTarget.Add(w.getLockKey())
			runWorker(w, minIOWorker.endWorker, minIOWorker.networkManager.CreateEndpoint())
		} else { //or enqueue
			minIOWorker.workQueue.Add(w)
		}

	}
	return nil
}

// remove the lock and trigger the first work possible
func (minIOWorker *MinIOWorker) handleEnd(end string) {
	println("end", end)
	minIOWorker.workingTarget.Remove(end)
	minIOWorker.currentConcurentWorker--

	for w, err := minIOWorker.workQueue.getNext(); err == nil; w, err = minIOWorker.workQueue.getNext() {
		if !minIOWorker.workingTarget.Contains(w.getLockKey()) {
			minIOWorker.currentConcurentWorker++
			minIOWorker.workingTarget.Add(w.getLockKey())
			runWorker(w, minIOWorker.endWorker, minIOWorker.networkManager.CreateEndpoint())
		}
	}
}

func getTempdirectory() string {
	uniqueName := uuid.New()
	tempDir := os.TempDir()
	if strings.HasSuffix(tempDir, "/") {
		tempDir = tempDir + uniqueName.String()
	} else {
		tempDir = tempDir + "/" + uniqueName.String()
	}
	tempDir += "/"
	os.Mkdir(tempDir, 0770)
	return tempDir
}

// For debug purpose
func pretyprint(o interface{}) string {
	bytes, _ := json.MarshalIndent(o, "", "  ")
	return string(bytes)
}
