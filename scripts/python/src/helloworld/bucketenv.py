#!/usr/bin/env python3
"""BucketEnv.

Usage:
  bucketenv.py <aws_account_id> <aws_region>
  bucketenv.py (-h | --help)

Commands:
   aws_account_id           aws account id
   aws_region               aws region

"""

from docopt import docopt
from troposphere import Template
from helloworld.common import CreateStack, WormHoleCredentials
from helloworld.bucket import BucketStack
import sys
import boto3
import botocore
from botocore.client import ClientError
import json


class EnvironmentBuilder(object):
                
    def buildEnvironment(self, createStack):   
        print(" calling create stack")        
        if createStack.checkIfStackExist():
            createStack.updateStack()
        else:
            self.stackId = createStack.createStack()
                
def main(aws_account_id, aws_region):
    
    wormCredentials = WormHoleCredentials(aws_account_id)
    wormHoleCredentials = wormCredentials.credentials()
    
    t = Template(Description="Bucket to hold the cold to deploy")
    t.add_version("2010-09-09")
    bucket_stack = BucketStack()

    bucket_stack.generate(t)
    parameters = [{
                    'ParameterKey': 'LambdaBucket',
                    'ParameterValue': 'go-lambda-hello-world'
                 }]
    stack_creator = CreateStack("go-lambda-deploy-bucket", 
                                t.to_json(), 
                                aws_region, 
                                parameters,
                                wormHoleCredentials)
    
    environmentBuilder = EnvironmentBuilder()
    environmentBuilder.buildEnvironment(stack_creator)
    
if __name__ == '__main__':
    arguments = docopt(__doc__, options_first=True)

    aws_account_id = arguments['<aws_account_id>']
    aws_region = arguments['<aws_region>']
    
    main(aws_account_id, aws_region)
    
    
