# -*- coding: utf-8 -*-
import boto3
import botocore
import pycurl
import getpass
import io
import json
from io import StringIO


class WormHoleCredentials:
    def __init__(self, accountId):
        self.accountId = accountId

    def credentials(self):
        
        #if getpass.getuser() == 'vagrant':
        #    cert_location = '/etc/pki/tls/certs/dev.bbc.co.uk.pem'
        #    cert_key_location = '/etc/pki/tls/private/dev.bbc.co.uk.key'
        #else:
        #    cert_location = '/etc/pki/tls/certs/client.crt'
        #    cert_key_location = '/etc/pki/tls/private/client.key'
        cert_location = "/etc/pki/certificate.pem"
        e = io.BytesIO()
        c = pycurl.Curl()
        c.setopt(c.URL, 'https://wormhole.api.bbci.co.uk/account/' + self.accountId + '/credentials')
        #c.setopt(pycurl.SSL_VERIFYPEER, 1)
        c.setopt(pycurl.SSL_VERIFYPEER, False)
        c.setopt(pycurl.SSL_VERIFYHOST, 2)
        c.setopt(c.WRITEFUNCTION, e.write)
        c.setopt(c.SSLCERT, cert_location)
        #c.setopt(c.SSLCERTPASSWD, '')
        #c.setopt(c.SSLKEY, cert_key_location)
        c.perform()
        c.close()

        contents = json.loads(e.getvalue().decode('UTF-8'))
        return contents

class CreateStack:
    
    def __init__(self, stackName, template,region, parameters, wormHoleCredentials):
    
        self.wormHoleCredentials = wormHoleCredentials
        self.capabilities = ['CAPABILITY_NAMED_IAM']
        self.template = template
        self.stackName = stackName
        self.region = region
        self.parameters = parameters
    
    def checkIfStackExist(self):
        client = boto3.client('cloudformation', region_name=self.region,
                                aws_access_key_id=self.wormHoleCredentials['accessKeyId'],
                                aws_secret_access_key=self.wormHoleCredentials['secretAccessKey'],
                                aws_session_token=self.wormHoleCredentials['sessionToken'])

        try:
            response = client.describe_stacks(StackName=self.stackName)
            return True
        
        except botocore.exceptions.ClientError as ce:
            return False
        
    def createStack(self):
        client = boto3.client('cloudformation', region_name=self.region,
                                aws_access_key_id=self.wormHoleCredentials['accessKeyId'],
                                aws_secret_access_key=self.wormHoleCredentials['secretAccessKey'],
                                aws_session_token=self.wormHoleCredentials['sessionToken'])
        try:
            client.validate_template(TemplateBody=self.template)

            response = client.create_stack(
                        StackName=self.stackName,
                        TemplateBody=self.template,
                        Parameters= self.parameters,
                        Capabilities= self.capabilities,
                        OnFailure='ROLLBACK'
                    )
            self.stackId = response['StackId']
            print("waiting for stack_create_complete [" + self.stackId +"]")
            waiter = client.get_waiter('stack_create_complete')
            waiter.wait(StackName=self.stackId)
            print("stack creation completed")
            
        except botocore.exceptions.ClientError as e:
            print(str(e))
            return None
        except botocore.exceptions.WaiterError as e:
            response = client.describe_stacks(StackName=stackId)
            print(str(response))
            return None
        
        return self.stackId
    
    def updateStack(self):
        client = boto3.client('cloudformation', region_name=self.region,
                                aws_access_key_id=self.wormHoleCredentials['accessKeyId'],
                                aws_secret_access_key=self.wormHoleCredentials['secretAccessKey'],
                                aws_session_token=self.wormHoleCredentials['sessionToken'],)
        try:
            response = client.update_stack(
                    StackName=self.stackName,
                    StackPolicyDuringUpdateBody=self.template,
                    UsePreviousTemplate=True,
                    Parameters = self.parameters,
                    Capabilities= self.capabilities
                    )
            self.stackId = response['StackId']
            print("waiting for stack_update_complete [" + self.stackId + "]")
            waiter = client.get_waiter('stack_update_complete')
            waiter.wait(StackName=self.stackId)
            print("stack update completed")

        except botocore.exceptions.ClientError as e:
            error = str(e)
            # based on comments from here https://github.com/hashicorp/terraform/issues/5653
            if error.find('No updates are to be performed') != -1:
                print("Nothing to update")
            else:
                print("something went wrong")
                print(error)
            return None
        except botocore.exceptions.WaiterError as e:

            if 'stackId' in locals() or 'stackId' in globals():
                response = client.describe_stacks(StackName=stackId)
                print(str(response))
            else:
                print("Stach was not created")    
            return None
        
        return self.stackId

    

