package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func messagePoller(svc *sqs.SQS, resultURL *sqs.GetQueueUrlOutput, messages chan<- string) {
	for {
		messages <- getMessage(svc, resultURL)
	}
}

func getMessage(svc *sqs.SQS, resultURL *sqs.GetQueueUrlOutput) string {
	// Receive a message from the SQS queue with long polling enabled.
	result, _ := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: resultURL.QueueUrl,
		AttributeNames: aws.StringSlice([]string{
			"SentTimestamp",
		}),
		MaxNumberOfMessages: aws.Int64(1),
		MessageAttributeNames: aws.StringSlice([]string{
			"All",
		}),
		VisibilityTimeout: aws.Int64(60),
		WaitTimeSeconds:   aws.Int64(1),
	})
	if len(result.Messages) > 0 {
		fmt.Printf("Received %d messages.\n", len(result.Messages))
		//fmt.Printf("%T\n", result.Messages[0])
		var message = *(result.Messages[0]).Body
		fmt.Println(message)
		_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      resultURL.QueueUrl,
			ReceiptHandle: result.Messages[0].ReceiptHandle,
		})
		if err != nil {
			fmt.Println("Delete Error", err)
		}
		return message
	}
	return ""
}

