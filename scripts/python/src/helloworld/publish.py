#!/usr/bin/env python3
"""Publish.

Usage:
  publish.py <aws_account_id> <aws_region>
  publish.py (-h | --help)


Commands:
   aws_account_id           aws account id
   aws_region               aws region
"""
import boto3
from docopt import docopt

from helloworld.common import WormHoleCredentials

class Publish:
    
    def __init__(self, region, wormHoleCredentials):
        self.region = region
        self.wormHoleCredentials = wormHoleCredentials

    def publish(self):
        sqs = boto3.resource('sqs', region_name=self.region,
                                aws_access_key_id=self.wormHoleCredentials['accessKeyId'],
                                aws_secret_access_key=self.wormHoleCredentials['secretAccessKey'],
                                aws_session_token=self.wormHoleCredentials['sessionToken'])


        # Get the queue
        queue = sqs.get_queue_by_name(QueueName='golambdaqueue')

        # Create a new message
        response = queue.send_message(MessageBody='ding dong')

        print(response)
        print(response.get('MessageId'))
        print(response.get('MD5OfMessageBody'))

    
if __name__ == '__main__':
    arguments = docopt(__doc__, options_first=True)

    aws_account_id = arguments['<aws_account_id>']
    aws_region = arguments['<aws_region>']
    
    wc = WormHoleCredentials(aws_account_id)
    wormHoleCredentials = wc.credentials()
    publish = Publish(aws_region, wormHoleCredentials)
    publish.publish()
    