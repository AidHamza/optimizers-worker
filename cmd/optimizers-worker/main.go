package main

import (
	//"os"
	"fmt"
	"bytes"
	//"time"
	"github.com/nsqio/go-nsq"
	//"github.com/spf13/afero"

	"github.com/AidHamza/optimizers-worker/pkg/storage"
	"github.com/AidHamza/optimizers-worker/pkg/messaging"
	"github.com/AidHamza/optimizers-worker/pkg/pool"
	"github.com/AidHamza/optimizers-worker/pkg/command"
	"github.com/AidHamza/optimizers-worker/pkg/operation"
	"github.com/AidHamza/optimizers-worker/pkg/config"
	"github.com/AidHamza/optimizers-worker/pkg/log"
)

const POOL_MAX_WORKERS int = 4

var STORAGE_BUCKETS = map[operation.Operation_FileType]string{
	operation.Operation_JPEG : "jpeg",
	operation.Operation_PNG : "png",
}

var OPTIMIZER_BUCKETS = map[operation.Operation_FileType]string{
	operation.Operation_JPEG : "optimized-jpeg",
	operation.Operation_PNG : "optimized-png",
}

// In-Memory File system holding directory
// Where we will execute temporary operations
// on files (Download / Optimize)
//var AppFs = afero.NewMemMapFs()
//var OSFs = afero.NewOsFs()
//var UFS = afero.NewCacheOnReadFs(OSFs, AppFs, 100 * time.Second)

func loop(messageIn chan *nsq.Message) {
	storageClient := storage.NewClient()

	pool := pool.New(POOL_MAX_WORKERS)

    for msg := range messageIn {
        operation, err := operation.LoadOperation(msg.Body)
        if err != nil {
        	fmt.Printf("Error in Unmarshal %+v", err)
        }

        fileName := fmt.Sprintf("%d/%s", operation.Id, operation.File)
        fileReader, err := storageClient.GetObject(STORAGE_BUCKETS[operation.Type], fileName)
		if err != nil {
			log.Logger.Error("Cannot load the file", "FILE_DOWNLOAD_FAILED", err.Error())
		}
		defer fileReader.Close()

		/*err = afero.WriteReader(UFS, "./tmp/" + fileName, fileReader)
		if err != nil {
			log.Logger.Error("Cannot save the file", "FILE_SAVE_FAILED", err.Error())
		}*/

		/*if _, err = os.Stat("./tmp/" + fileName); os.IsNotExist(err) {
			log.Logger.Error("File not exists", "FILE_NOT_EXISTS", err.Error())
		}*/
		fileInfo, _ := fileReader.Stat()
		fileBuffer := make([]byte, fileInfo.Size)
		fileReader.Read(fileBuffer)

		optimizeCMD := func() {
			handler := command.NewHandler()

			// DEBUG MODE ON
			handler.EnableDebug(false)

			imgBuf, err := handler.RunCommand(operation.Command[0].Command, operation.Command[0].Flags, fileBuffer)
			fileBuffer = nil
			if err != nil {
				log.Logger.Error("Cannot optimize the file", "FILE_OPTIMIZE_FAILED", err.Error())
			}

			b := bytes.NewReader(imgBuf)
			defer b.Reset(imgBuf)
			imgBuf = nil

			err = storageClient.PutObject(b, OPTIMIZER_BUCKETS[operation.Type], fileName, fileInfo.ContentType)
			if err != nil {
				log.Logger.Error("Cannot store optimized file", "FILE_STORE_FAILED", err.Error())
			}

			fmt.Println("Operation:", operation.Type)
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

	/*exists, err := afero.DirExists(UFS, "./tmp")
	if err != nil {
		log.Logger.Error("Cannot stat operations directory", "TMP_DIR_STAT", err.Error())
	}

	if exists == false {
		err = UFS.Mkdir("./tmp", 0755)
		if err != nil {
			log.Logger.Error("Cannot create operations directory", "TMP_DIR_STAT_CREATE", err.Error())
		}
	}*/

	messageIn := make(chan *nsq.Message)
	c := messaging.NewConsumer("jpeg", "jpeg_optimization")

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