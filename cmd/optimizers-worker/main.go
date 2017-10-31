package main

import (
	"fmt"
	"log"
	"github.com/nsqio/go-nsq"
	"github.com/AidHamza/optimizers-worker/pkg/messaging"
	"github.com/AidHamza/optimizers-worker/pkg/pool"
	"github.com/AidHamza/optimizers-worker/pkg/command"
)

const POOL_MAX_WORKERS int = 4

func loop(messageIn chan *nsq.Message) {
	args := []string{"google.com", "-c", "2"}
	ping := func() {
		handler := command.NewHandler()
		handler.RunCommand("ping", args)
	}

	pool := pool.New(POOL_MAX_WORKERS)

    for msg := range messageIn {
    	fmt.Printf("Loop")
    	pool.Schedule(ping)
        log.Println(string(msg.Body))
        msg.Finish()
    }
}

func main() {
	messageIn := make(chan *nsq.Message)
	c := messaging.NewConsumer("jpeg", "jpeg_optimization")

	c.Set("nsqd", ":32771")
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