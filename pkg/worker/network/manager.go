package network

import (
	"context"
	"errors"

	"github.com/minio/minio-go/v7"
)

//NetworkManager this object is responsible of the network requests
type NetworkManager struct {
	networkQueue            *requestQueue
	maxConcurentAccess      int
	currentConcurrentAccess int
	endpoint                *minio.Client
	endRequestChan          chan interface{}
	requestChan             chan networkRequest
	ctx                     context.Context
}

//NewNetworkManager instantiate a new networkManager and start it
func NewNetworkManager(ctx context.Context, maxConcurentAccess int, endpoint *minio.Client) *NetworkManager {
	manager := &NetworkManager{
		networkQueue:            &requestQueue{},
		maxConcurentAccess:      maxConcurentAccess,
		currentConcurrentAccess: 0,
		endpoint:                endpoint,
		endRequestChan:          make(chan interface{}),
		requestChan:             make(chan networkRequest),
		ctx:                     ctx,
	}
	go manager.run()
	return manager
}

//Run main loop of the network manager
func (manager *NetworkManager) run() {
	for {
		select {
		case <-manager.ctx.Done():
			return
		case req := <-manager.requestChan:
			manager.handleRequest(req)
		case <-manager.endRequestChan:
			manager.handleEndRequest()
		}
	}
}

func (manager *NetworkManager) handleRequest(req networkRequest) {
	if manager.currentConcurrentAccess < manager.maxConcurentAccess {
		manager.currentConcurrentAccess++
		handleRequest(manager.endpoint, req, manager.endRequestChan)
	} else {
		manager.networkQueue.Add(req)
	}
}

func (manager *NetworkManager) handleEndRequest() {
	req, err := manager.networkQueue.Pop()
	if err != nil {
		manager.currentConcurrentAccess--
	} else {
		handleRequest(manager.endpoint, req, manager.endRequestChan)
	}
}

//CreateEndpoint create an endpoint to interact with this manager
func (manager *NetworkManager) CreateEndpoint() NetworkEndpoint {
	return NetworkEndpoint{
		requestChan:    manager.requestChan,
		notifyDownChan: make(chan error),
		notifyUpChan:   make(chan error),
		endAllUpChan:   make(chan []error),
		count:          0,
		uploadReady:    true,
	}

}

//NetworkEndpoint endpoint to interact with the manager
type NetworkEndpoint struct {
	requestChan    chan networkRequest
	notifyDownChan chan error
	notifyUpChan   chan error
	endAllUpChan   chan []error
	count          int
	uploadReady    bool
}

func (endpoint *NetworkEndpoint) notifyEnd() {
	defer func() { endpoint.uploadReady = true }()
	count := endpoint.count
	errs := make([]error, 0, count)
	for err := range endpoint.notifyUpChan {
		if err != nil {
			errs = append(errs, err)
		}
		endpoint.count-- //may run to concurent access
		count--
		if count == 0 {
			if len(errs) > 0 {
				endpoint.endAllUpChan <- errs
				return
			}
			endpoint.endAllUpChan <- nil
			return
		}
	}
}

//Download ask to the manager to download a file and wait util it's completion
func (endpoint *NetworkEndpoint) Download(bucket string, remotePath string, localpath string) error {
	endpoint.requestChan <- networkRequest{
		bucket:      bucket,
		remotepath:  remotePath,
		localpath:   localpath,
		requestType: download,
		notifChan:   endpoint.notifyDownChan,
	}
	//wait for download complete
	return <-endpoint.notifyDownChan
}

//Upload ask to the manager to upload a file, do not woit for it's completion
func (endpoint *NetworkEndpoint) Upload(bucket string, remotePath string, localpath string) error {
	if !endpoint.uploadReady {
		return errors.New("upload not ready")
	}
	endpoint.count++ //may run to concurent access
	endpoint.requestChan <- networkRequest{
		bucket:      bucket,
		remotepath:  remotePath,
		localpath:   localpath,
		notifChan:   endpoint.notifyUpChan,
		requestType: upload,
	}
	return nil
}

//WaitAll await for all the request to be completed
func (endpoint *NetworkEndpoint) WaitAll() []error {
	endpoint.uploadReady = false
	go endpoint.notifyEnd()
	if !endpoint.uploadReady {
		return nil
	}
	return <-endpoint.endAllUpChan
}
