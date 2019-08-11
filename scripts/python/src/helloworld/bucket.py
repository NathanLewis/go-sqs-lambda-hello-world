
from troposphere import Output, Ref, Template
from troposphere.s3 import Bucket, PublicRead
from troposphere import Template,Ref,Parameter,Output, Export
from troposphere.s3 import Rules as S3Key,Bucket

class BucketStack:

    def generate(self, template):

        lambda_bucket = template.add_parameter(Parameter(
            "LambdaBucket",
            Type="String",
            Default="go-lambda-hello-world",
            Description="The S3 Bucket that contains the zip to bootstrap your "
                "lambda function"
        ))

        s3bucket = template.add_resource(Bucket(
                "LambdaDeployBucket",
                BucketName=Ref(lambda_bucket), 
                AccessControl=PublicRead,)
        )
        
        template.add_output(Output(
            "BucketName",
            Value=Ref(s3bucket),
            Export=Export("go-lambda-deploy-bucket"),
            Description="Name of S3 bucket to hold website content"
        ))

if __name__ == '__main__':
    t = Template(Description="S3 bucket used to deploy the lambda")
    t.add_version("2010-09-09")
    deploy_bucket = Bucket()
    deploy_bucket.generate(t)
    print(t.to_json())        

