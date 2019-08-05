// This code does not even compile yet

// Snippets taken from https://medium.com/@marcioghiraldelli/elegant-use-of-golang-channels-with-aws-sqs-dad20cd59f34

func pollSqs(chn chan<- *sqs.Message) {
	
  for {
    output, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
      QueueUrl:            &config.StackNotificationSqsUrl,
      MaxNumberOfMessages: aws.Int64(sqsMaxMessages),
      WaitTimeSeconds:     aws.Int64(sqsPollWaitSeconds),
    })

    if err != nil {
      log.Errorf("failed to fetch sqs message %v", err)
    }

    for _, message := range output.Messages {
      chn <- message
    }

  }
	
}

func main() {
	
  chnMessages := make(chan *sqs.Message, sqsMaxMessages)
  go pollSqs(chnMessages)

  log.Infof("Listening on stack queue: %s", config.StackNotificationSqsUrl)

  for message := range chnMessages {
    handleMessage(message)
    deleteMessage(message)
  }

}
