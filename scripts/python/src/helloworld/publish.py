import boto3

# Get the service resource
sqs = boto3.resource('sqs')

# Get the queue
queue = sqs.get_queue_by_name(QueueName='golambdaqueue')

# Create a new message
response = queue.send_message(MessageBody='ding dong')

print(response)
print(response.get('MessageId'))
print(response.get('MD5OfMessageBody'))
