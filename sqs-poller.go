package main

// Code taken from https://h2ik.co/2017/01/28/go-sqs-poller/

import (
	"fmt"

	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/h2ik/go-sqs-poller/worker"
)

var (
	aws_access_key string
	aws_secret_key string
	aws_region     string
)

func main() {

	// create a config
	aws_config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(aws_access_key, aws_secret_key, ""),
		Region:      aws.String(aws_region),
	}

	// create the new client with the aws_config passed in
	svc, url := worker.NewSQSClient("go-webhook-queue-test", aws_config)
	// set the queue url
	worker.QueueURL = url
	// start the worker
	worker.Start(svc, worker.HandlerFunc(func(msg *sqs.Message) error {
		fmt.Println(aws.StringValue(msg.Body))
		return nil
	}))
}

