package network

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
)

func handleRequest(endpoint *minio.Client, req networkRequest, endChan chan interface{}) {
	handler := requestHandler{
		endpoint: endpoint,
		req:      req,
		endChan:  endChan,
	}
	go handler.Run()
}

type requestHandler struct {
	endpoint *minio.Client
	req      networkRequest
	endChan  chan interface{}
}

func (handler *requestHandler) Run() {
	var err error
	if handler.req.requestType == download {
		err = handler.handleDownload()
	} else {
		err = handler.handleUpload()
	}
	if err != nil {
		log.Println("error while", handler.req.requestType, handler.req.localpath, handler.req.bucket, handler.req.remotepath, err)
	}

	handler.endChan <- struct{}{}
	//non blocking publish
	select {
	case handler.req.notifChan <- err:
	default:
	}

}

func (handler *requestHandler) handleDownload() error {
	obj, err := handler.endpoint.GetObject(context.Background(), handler.req.bucket, handler.req.remotepath, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	file, err := os.OpenFile(handler.req.localpath, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, obj)
	if err != nil {
		return err
	}
	return nil
}

func (handler *requestHandler) handleUpload() error {
	// make sure bucket exists
	ok, err := handler.endpoint.BucketExists(context.Background(), handler.req.bucket)
	if err != nil {
		return err
	}
	if !ok {
		if err = handler.endpoint.MakeBucket(context.Background(), handler.req.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	//upload file
	//get file size
	stat, err := os.Stat(handler.req.localpath)

	continueErr := errors.New("")
	count := 0
	for ; continueErr != nil && count < 3; count++ {
		//get reader
		reader, err := os.Open(handler.req.localpath)
		if err != nil {
			return err
		}
		//write on minio
		_, continueErr = handler.endpoint.PutObject(context.Background(), handler.req.bucket, handler.req.remotepath, reader, stat.Size(), minio.PutObjectOptions{})
	}
	return continueErr
}
