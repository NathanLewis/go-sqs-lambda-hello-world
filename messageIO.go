package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func messagePoller(svc *sqs.SQS, resultURL *sqs.GetQueueUrlOutput, messages chan<- string) {
	for {
		message := getMessage(svc, resultURL)
		if 0 == len(message) {
			continue
		}
		messages <- message
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

func messageSender(svc *sqs.SQS, badMesgQueueURL *sqs.GetQueueUrlOutput, messages chan string) {
	for {
		message := <- messages
		_, err := svc.SendMessage(&sqs.SendMessageInput{
			DelaySeconds: aws.Int64(0),
			MessageBody:  aws.String(message),
			QueueUrl:     badMesgQueueURL.QueueUrl,
		})
		if err != nil {
			fmt.Printf("Unable to send to Queue %s\n", badMesgQueueURL.QueueUrl)
		}
	}
}

