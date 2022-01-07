package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalf("error: failed to load configuration, %v", err)
	}
	cli := sqs.NewFromConfig(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("name: %s", os.Getenv("QUEUE_NAME"))
	log.Printf("owner id: %s", os.Getenv("QUEUE_OWNER_ID"))

	ui := &sqs.GetQueueUrlInput{
		QueueName:              aws.String(os.Getenv("QUEUE_NAME")),
		QueueOwnerAWSAccountId: aws.String(os.Getenv("QUEUE_OWNER_ID")),
	}

	output, err := cli.GetQueueUrl(ctx, ui)
	if err != nil {
		log.Fatalf("error: failed to get queue url, %v", err)
	}
	url := output.QueueUrl

	config := Config{
		QueueUrl:    url,
		MaxWorker:   1,
		MaxMsg:      10,
		WaitTimeout: 20,
	}

	NewConsumer(cli, config).Start(ctx)
}
