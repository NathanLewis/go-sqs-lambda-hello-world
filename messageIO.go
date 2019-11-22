package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strings"
)

func messagePoller(svc *sqs.SQS, url *sqs.GetQueueUrlOutput, messages chan<- string) {
	for {
		message := getMessage(svc, url)
		if 0 == len(message) {
			continue
		}
		messages <- message
	}
}

func getMessage(svc *sqs.SQS, url *sqs.GetQueueUrlOutput) string {
	// Receive a message from the SQS queue with long polling enabled.
	result, _ := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: url.QueueUrl,
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
			QueueUrl:      url.QueueUrl,
			ReceiptHandle: result.Messages[0].ReceiptHandle,
		})
		if err != nil {
			fmt.Println("Delete Error", err)
		}
		return message
	}
	return ""
}

func sqsSender(svc *sqs.SQS, url *sqs.GetQueueUrlOutput, messages chan string) {
	for {
		message := <- messages
		_, err := svc.SendMessage(&sqs.SendMessageInput{
			DelaySeconds: aws.Int64(0),
			MessageBody:  aws.String(message),
			QueueUrl:     url.QueueUrl,
		})
		if err != nil {
			fmt.Printf("Unable to send to Queue %s\n", url.QueueUrl)
		}
	}
}


func snsSender(sess *session.Session, topicArn string, messages <-chan string) {
	client := sns.New(sess)
	for {
		message := <- messages
		result, err := client.Publish(
			&sns.PublishInput{Message: aws.String(message),
				TopicArn: aws.String(topicArn),
			})
		if err != nil {
			fmt.Println("Publish error:", err)
		}
		fmt.Println(result)
	}
}


func findQueueUrl(svc *sqs.SQS, queueName string) *sqs.GetQueueUrlOutput {
	// I would have thought that if it began with https I could just return it
	if strings.HasPrefix(queueName, "https:") {
		var queueUrl sqs.GetQueueUrlOutput
		queueUrl.SetQueueUrl(queueName)
		return &queueUrl
	}
	/**/
	if strings.HasPrefix(queueName,"arn:") {
		slice := strings.SplitN(queueName, ":", -1)
		queueName = slice[len(slice)-1:][0]
	}
	queueURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			exitErrorf("Unable to find queue %q.", queueName)
		}
		exitErrorf("Unable to queue %q, %v.", queueName, err)
	}
	return queueURL
}

func setupSession() (*sqs.SQS, *session.Session) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	if err != nil {
		exitErrorf("Unable to setup Session")
	}

	// Create a SQS service client.
	svc := sqs.New(sess)
	return svc, sess
}

