package messaging_test

import (
	"bytes"
	"github.com/bmizerany/assert"
	"github.com/nsqio/go-nsq"
	"github.com/AidHamza/optimizers-worker/pkg/messaging"
	"log"
	"testing"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestConsumer_Start_ConfigErrors(t *testing.T) {
	c := messaging.NewConsumer("events", "ingestion")

	c.Set("nsqd", ":32769")
	c.Set("concurrency", 5)
	c.Set("max_attempts", 10)
	c.Set("max_in_flight", "oh noes")
	c.Set("default_requeue_delay", "15s")

	err := c.Start(nil)
	assert.Equal(t, `failed to coerce option max_in_flight (oh noes) - strconv.ParseInt: parsing "oh noes": invalid syntax`, err.Error())
}

func TestConsumer_Start_ConfigMissingDaemon(t *testing.T) {
	c := messaging.NewConsumer("events", "ingestion")

	c.Set("concurrency", 5)
	c.Set("max_attempts", 10)

	err := c.Start(nil)
	assert.Equal(t, `at least one "nsqd" or "nsqlookupd" address must be configured`, err.Error())
}

func TestConsumer_Start_Handler(t *testing.T) {
	done := make(chan bool)
	b := new(bytes.Buffer)
	l := log.New(b, "", 0)

	c := messaging.NewConsumer("events", "ingestion")
	c.SetLogger(l, nsq.LogLevelDebug)

	c.Set("nsqd", ":32769")
	c.Set("nsqds", []interface{}{":32769"})
	c.Set("concurrency", 5)
	c.Set("max_attempts", 10)
	c.Set("max_in_flight", 150)
	c.Set("default_requeue_delay", "15s")

	err := c.Start(nsq.HandlerFunc(func(msg *nsq.Message) error {
		done <- true
		return nil
	}))

	assert.Equal(t, nil, err)

	go func() {
		p, err := nsq.NewProducer(":32769", nsq.NewConfig())
		check(err)
		p.Publish("events", []byte("hello"))
	}()

	<-done
	assert.Equal(t, nil, c.Stop())
}