package main

import (
	"context"
	"errors"
	"fmt"
    "os"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strconv"
	"time"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	if len(sqsEvent.Records) == 0 {
		return errors.New("No SQS message passed to function")
	}
	// Initialize a session in eu-west-1 that the SDK will use to load
	sqs, sess := setupSession()
	badMesgQueueURL := os.Getenv("BAD_MESSAGE_QUEUE_URL")
	snsTopicArn := os.Getenv("ISPY_TOPIC_ARN")

	badMesgQueueChannel := make(chan string)
	snsChannel := make(chan map[string]interface{})
	go sqsSender(sqs, &badMesgQueueURL, badMesgQueueChannel)
	go snsSender(sess, snsTopicArn, snsChannel)

	for _, msg := range sqsEvent.Records {
		fmt.Printf("Got SQS message %q with body %q\n", msg.MessageId, msg.Body)
        message := msg.Body
		var asset Asset
		var snsMesgMap map[string]interface{}
        err := asset.readFromString(message)
        if err == nil {
            snsMesgMap = asset.toMap()
            snsMesgMap["event_name"] = "go.sqs.message.status.ok"
        } else {
            fmt.Printf("Bad Message: %s\n", message)
            badMesgQueueChannel <- message
            snsMesgMap = make(map[string]interface{})
            snsMesgMap["event_name"] = "go.sqs.message.status.badmessage"
            snsMesgMap["bad_message"] = message
        }
        // Add whatever fields you want / need to the sns message
        snsMesgMap["event_timestamp"] = strconv.FormatInt(int64(time.Now().Unix()), 10)

        snsChannel <- snsMesgMap
	}

	//return fmt.Sprintf("Finished Handling %d events!", len(sqsEvent.Records)), nil
	return nil
}
