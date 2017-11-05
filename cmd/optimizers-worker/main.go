package main

import (
	"fmt"
	"bytes"
	"github.com/nsqio/go-nsq"

	"github.com/AidHamza/optimizers-worker/pkg/storage"
	"github.com/AidHamza/optimizers-worker/pkg/messaging"
	"github.com/AidHamza/optimizers-worker/pkg/pool"
	"github.com/AidHamza/optimizers-worker/pkg/command"
	"github.com/AidHamza/optimizers-worker/pkg/operation"
	"github.com/AidHamza/optimizers-worker/pkg/config"
	"github.com/AidHamza/optimizers-worker/pkg/log"
)

const POOL_MAX_WORKERS int = 4
const RESULTS_QUEUE_TOPIC string = "operationResult"

var STORAGE_BUCKETS = map[operation.Operation_FileType]string{
	operation.Operation_JPEG : "jpeg",
	operation.Operation_PNG : "png",
}

var OPTIMIZER_BUCKETS = map[operation.Operation_FileType]string{
	operation.Operation_JPEG : "optimized-jpeg",
	operation.Operation_PNG : "optimized-png",
}

func loop(messageIn chan *nsq.Message) {
	storageClient := storage.NewClient()

	producerClient, err := messaging.NewProducer(config.App.Messaging.Host, config.App.Messaging.Port)
	if err != nil {
		log.Logger.Error("Cannot initialize producer", "PRODUCER_INIT_FAILED", err.Error())
	}

	pool := pool.New(POOL_MAX_WORKERS)

    for msg := range messageIn {
        operationBuf, err := operation.LoadOperation(msg.Body)
        defer operationBuf.Reset()
        if err != nil {
        	fmt.Printf("Error in Unmarshal %+v", err)
        }

		optimizeCMD := func() {

			fileName := fmt.Sprintf("%d/%s", operationBuf.Id, operationBuf.File)
	        fileReader, err := storageClient.GetObject(STORAGE_BUCKETS[operationBuf.Type], fileName)
			if err != nil {
				log.Logger.Error("Cannot load the file", "FILE_DOWNLOAD_FAILED", err.Error())
			}
			defer fileReader.Close()

			fileInfo, _ := fileReader.Stat()
			fileBuffer := make([]byte, fileInfo.Size)
			fileReader.Read(fileBuffer)
		
			handler := command.NewHandler()

			// DEBUG MODE ON
			handler.EnableDebug(false)

			imgBuf, err := handler.RunCommand(operationBuf.Command[0].Command, operationBuf.Command[0].Flags, fileBuffer)
			fileBuffer = nil
			if err != nil {
				log.Logger.Error("Cannot optimize the file", "FILE_OPTIMIZE_FAILED", err.Error())
			}

			b := bytes.NewReader(imgBuf)
			defer b.Reset(imgBuf)
			imgBuf = nil

			err = storageClient.PutObject(b, OPTIMIZER_BUCKETS[operationBuf.Type], fileName, fileInfo.ContentType)
			if err != nil {
				log.Logger.Error("Cannot store optimized file", "FILE_STORE_FAILED", err.Error())
			}

			operationResult, err := operation.NewResult(operationBuf.Id, operationBuf.File, operation.Result_SUCCESS)
			if err != nil {
				log.Logger.Error("Failed to marshal operation result to Protobuf", "RESULT_MARSHAL_FAILED", err.Error())
			}

			err = producerClient.PublishMessage(RESULTS_QUEUE_TOPIC, operationResult)
			operationResult = nil
			if err != nil {
				log.Logger.Error("Unable to queue result", "RESULT_QUEUE_FAILED", err.Error())
			}
		}
		pool.Schedule(optimizeCMD)
        msg.Finish()
    }
}

func main() {
	err := config.App.Init()
	if err != nil {
		log.Logger.Error("Cannot load app configuration", "VIPER_ERROR", err.Error())
	}

	messageIn := make(chan *nsq.Message)
	c := messaging.NewConsumer(config.App.Messaging.Topic, "optimization")

	c.Set("nsqd", ":" + config.App.Messaging.Port)
	c.Set("concurrency", 15)
	c.Set("max_attempts", 10)
	c.Set("max_in_flight", 1000)
	c.Set("default_requeue_delay", "15s")

	c.Start(nsq.HandlerFunc(func(msg *nsq.Message) error {
		messageIn <- msg
		return nil
	}))

	forever := make(chan bool)
	go loop(messageIn)
	<-forever

	c.Stop()
	return
}