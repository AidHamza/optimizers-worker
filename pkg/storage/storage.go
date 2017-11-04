package storage

import (
	"fmt"
	"io"
	"github.com/minio/minio-go"
	"github.com/AidHamza/optimizers-worker/pkg/config"
	"github.com/AidHamza/optimizers-worker/pkg/log"
)

type Client interface {
	PutObject(io.Reader, string, string, string) error
	GetObject(string, string) (*minio.Object, error)
}

type client struct {
	API *minio.Client
}

func NewClient() Client {
	clientAPI, err := minio.New(
		fmt.Sprintf("%s:%d", config.App.Storage.Host, config.App.Storage.Port),
		config.App.Storage.AccessKey,
		config.App.Storage.SecretKey,
		config.App.Storage.TLS)

	if err != nil {
		panic(err.Error())
	}

	return &client{
		API: clientAPI,
	}
}

func (c *client) PutObject(fileReader io.Reader, bucket string, fileName string, fileType string) error {
	n, err := c.API.PutObject(bucket, fileName, fileReader, fileType)
	if err != nil {
		return err
	}

	log.Logger.Info("File stored", "FILENAME", fileName, "SIZE", n)

	return nil
}

func (c *client) GetObject(bucket string, fileName string) (*minio.Object, error) {
	imgObject, err := c.API.GetObject(bucket, fileName)
	if err != nil {
		return nil, err
	}

	log.Logger.Info("File downloaded", "FILENAME", fileName)

	return imgObject, nil
}

func (c *client) SaveObject(bucket string, fileName string, savePath string) error {
	err := c.API.FGetObject(bucket, fileName, savePath)
	if err != nil {
		return err
	}

	log.Logger.Info("File saved", "FILENAME", fileName, "PATH", savePath)

	return nil
}