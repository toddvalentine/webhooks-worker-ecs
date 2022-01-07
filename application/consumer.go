package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Config struct {
	QueueUrl    *string
	MaxWorker   int
	MaxMsg      int
	WaitTimeout int
}

type Consumer struct {
	client *sqs.Client
	config Config
}

func NewConsumer(client *sqs.Client, config Config) Consumer {
	return Consumer{
		client: client,
		config: config,
	}
}

func (c Consumer) Start(ctx context.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(c.config.MaxWorker)

	for i := 1; i <= c.config.MaxWorker; i++ {
		go c.worker(ctx, wg, i)
	}

	wg.Wait()
}

func (c Consumer) worker(ctx context.Context, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	log.Printf("info: worker %d started\n", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("info: worker %d stopped\n", id)
			return
		default:
		}

		min := &sqs.ReceiveMessageInput{
			QueueUrl:              c.config.QueueUrl,
			MaxNumberOfMessages:   int32(c.config.MaxMsg),
			WaitTimeSeconds:       int32(c.config.WaitTimeout),
			MessageAttributeNames: []string{"Endpoint", "Signature"},
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(min.WaitTimeSeconds+5))
		defer cancel()

		mout, err := c.client.ReceiveMessage(ctx, min)
		if err != nil {
			log.Printf("error: receive message call...: %v\n", err)
			continue
		}

		if len(mout.Messages) == 0 {
			continue
		}

		for _, msg := range mout.Messages {
			dmi := &sqs.DeleteMessageInput{
				QueueUrl:      c.config.QueueUrl,
				ReceiptHandle: msg.ReceiptHandle,
			}
			c.consume(ctx, msg, dmi)
		}
	}

}

func (c Consumer) consume(ctx context.Context, msg types.Message, dmi *sqs.DeleteMessageInput) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	svc := ServiceCaller{
		client: &http.Client{
			Timeout: time.Second * 3,
		},
	}

	endpoint := *msg.MessageAttributes["Endpoint"].StringValue
	signature := *msg.MessageAttributes["Signature"].StringValue
	body := *msg.Body

	status, err := svc.callService(ctx, endpoint, signature, body)
	if err != nil {
		log.Printf("error: failed to call service endpoint, %v", err)
		return
	}

	if status != 200 {
		log.Printf("error: did not receive 200 ok, %v", err)
		return
	}
	_, err = c.client.DeleteMessage(ctx, dmi)
	if err != nil {
		log.Printf("error: failed to delete message, %v", err)
	}

	log.Printf("info: message id %s deleted off queue", *msg.MessageId)
}
