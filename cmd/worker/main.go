package main

import (
	"context"
	"flag"
	"log"

	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/network"
	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/pipeline"

	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker"
	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/configuration"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	//read configuration
	config := flag.String("config", "config.json", "the path to the config file")

	globalConfig := configuration.LoadConfiguration(*config)

	//load pipelines
	pipelines, err := pipeline.Load(globalConfig.PipelineDirectory)
	if err != nil {
		log.Fatal("error while loading pipelines ", err)
	}

	//initialise the connection
	endpoint, err := minio.New(globalConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(globalConfig.AccessKeyID, globalConfig.SecretAccessKey, ""),
		Secure: globalConfig.UseSSl,
	})
	if err != nil {
		log.Fatal("error connecting to the minio instance", err)
	}
	eventCtx, eventStopTrigger := context.WithCancel(context.Background())
	eventChan := endpoint.ListenNotification(eventCtx, "", "", []string{
		"s3:ObjectCreated:*",
	})
	//initialise actors
	//NetworkManager
	networkContext, networkStopTrigger := context.WithCancel(context.Background())
	networkManager := network.NewNetworkManager(networkContext, globalConfig.MaxConcurentUploadFiles, endpoint)
	//minioWorker
	minioWorker := worker.NewMinIOWorker(eventChan, globalConfig.MaxConcurentWorker, pipelines, networkManager, globalConfig.ResultsDestination)
	//mainloop on events
	err = minioWorker.Run()

	if err != nil {
		log.Println(err) //do not fail here, gracefully stop the actors
	}

	//stop all actors
	eventStopTrigger()
	networkStopTrigger()
}
