#!/usr/bin/env python3
"""EnvironmentBuilder.

Usage:
  environmentbuilder.py <aws_account_id> <aws_region>
  environmentbuilder.py (-h | --help)


Commands:
   aws_account_id           aws account id
   aws_region               aws region
"""

from docopt import docopt
from troposphere import Template
from helloworld.common import CreateStack, WormHoleCredentials
from helloworld.function import LambdaStack
import sys
import boto3
import botocore
from botocore.client import ClientError
import json


class CreateEventSourceMapping:
    
    def __init__(self, region, wormCredentials):
        self.wormHoleCredentials = wormCredentials
        self.region = region
    def createMapping(self):
        cf = boto3.client('cloudformation', region_name=self.region,
                                aws_access_key_id=self.wormHoleCredentials['accessKeyId'],
                                aws_secret_access_key=self.wormHoleCredentials['secretAccessKey'],
                                aws_session_token=self.wormHoleCredentials['sessionToken'])


        ex = cf.list_exports()['Exports']

        for i in ex:
            if i['Name'] == 'lambda-go-hello-world-function':
                function_arn = i['Value']

            if i['Name'] == 'lambda-go-hello-world-queue':
                queue_arn = i['Value']
        

        print(function_arn)
        print("\n"+queue_arn)
        client = boto3.client('lambda', region_name=self.region,
            aws_access_key_id=self.wormHoleCredentials['accessKeyId'],
            aws_secret_access_key=self.wormHoleCredentials['secretAccessKey'],
            aws_session_token=self.wormHoleCredentials['sessionToken'])

        return client.create_event_source_mapping(
                    EventSourceArn=queue_arn,
                    FunctionName=function_arn,
                    Enabled=True,
                    BatchSize=1,)

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

    t = Template(Description="Go lambda Stack")
    t.add_version("2010-09-09")
    lambda_stack = LambdaStack()

    aws_lambda = lambda_stack.generate_template(t)
    parameters = [{
                    'ParameterKey': 'LambdaBucket',
                    'ParameterValue': 'go-lambda-hello-world'
                 },{
                    'ParameterKey': 'S3Key',
                    'ParameterValue': 'main.zip'
                },{
                    'ParameterKey': 'LambdaEnvAlias',
                    'ParameterValue': 'int'
                },{
                    'ParameterKey': 'LambdaMemorySize',
                    'ParameterValue': '128'
                },{
                    'ParameterKey': 'LambdaTimeout',
                    'ParameterValue': '300'
                },{
                    'ParameterKey': 'LambdaHandler',
                    'ParameterValue': 'main'
                }]
    stack_creator = CreateStack("go-lambda-sqs", 
                                t.to_json(), 
                                aws_region, 
                                parameters,
                                wormHoleCredentials)
    
    environmentBuilder = EnvironmentBuilder()
    environmentBuilder.buildEnvironment(stack_creator)

    createEventSourceMapping = CreateEventSourceMapping(aws_region, wormHoleCredentials)
    print(createEventSourceMapping.createMapping())
    
if __name__ == '__main__':
    arguments = docopt(__doc__, options_first=True)

    aws_account_id = arguments['<aws_account_id>']
    aws_region = arguments['<aws_region>']
    
    main(aws_account_id, aws_region)
    
    
